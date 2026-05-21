package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ehubgo/cache"
)

type WsHub struct {
	chans      map[string]chan []byte // driver user_id -> chan
	mu         sync.RWMutex
	redisStore cache.Store
}

type DriverNotification struct {
	TargetDriverIDs []string    `json:"target_driver_ids"` // Empty means broadcast to all
	Payload         interface{} `json:"payload"`
}

func NewWsHub(redisStore cache.Store) *WsHub {
	hub := &WsHub{
		chans:      make(map[string]chan []byte),
		redisStore: redisStore,
	}
	if redisStore != nil && redisStore.IsAvailable() {
		go hub.listenToRedis()
	}
	return hub
}

func (h *WsHub) listenToRedis() {
	ctx := context.Background()
	ch := h.redisStore.Subscribe(ctx, "driver_notifications")
	log.Println("[WsHub] Subscribed to Redis channel: driver_notifications")

	for msg := range ch {
		var notification DriverNotification
		if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
			log.Printf("[WsHub] Failed to unmarshal Redis message: %v", err)
			continue
		}

		// Deliver to local connections
		h.deliverLocally(notification)
	}
}

func (h *WsHub) deliverLocally(n DriverNotification) {
	b, err := json.Marshal(n.Payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(n.TargetDriverIDs) == 0 {
		// Broadcast to all locally connected drivers
		for _, ch := range h.chans {
			select {
			case ch <- b:
			default:
			}
		}
		return
	}

	// Targeted delivery
	for _, id := range n.TargetDriverIDs {
		if ch, ok := h.chans[id]; ok {
			select {
			case ch <- b:
			default:
			}
		}
	}
}

func (h *WsHub) Register(driverID string) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 10)
	h.chans[driverID] = ch
	return ch
}

func (h *WsHub) Unregister(driverID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.chans[driverID]; ok {
		close(ch)
		delete(h.chans, driverID)
	}
}

func (h *WsHub) SendToDriver(driverID string, v interface{}) error {
	notification := DriverNotification{
		TargetDriverIDs: []string{driverID},
		Payload:         v,
	}

	if h.redisStore != nil && h.redisStore.IsAvailable() {
		b, _ := json.Marshal(notification)
		return h.redisStore.Publish(context.Background(), "driver_notifications", b)
	}

	// Fallback to local delivery if Redis is down
	h.deliverLocally(notification)
	return nil
}

func (h *WsHub) BroadcastToDrivers(driverIDs []string, v interface{}) {
	notification := DriverNotification{
		TargetDriverIDs: driverIDs,
		Payload:         v,
	}

	if h.redisStore != nil && h.redisStore.IsAvailable() {
		b, _ := json.Marshal(notification)
		_ = h.redisStore.Publish(context.Background(), "driver_notifications", b)
		return
	}

	// Fallback to local delivery if Redis is down
	h.deliverLocally(notification)
}

// global hub instance (re-initialized in main)
var DriverHub *WsHub

// SSE helper to stream events to a response writer
func StreamEvents(w http.ResponseWriter, flusher http.Flusher, ch chan []byte, stop <-chan struct{}) {
	// send a ping event to confirm connection
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-stop:
			return
		case <-time.After(60 * time.Second):
			// heartbeat
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}

// DeclineStore keeps in-memory record of driver declines per order with TTL
type DeclineStore struct {
	mu sync.Mutex
	m  map[string]map[string]time.Time // orderID -> driverID -> timestamp
}

func NewDeclineStore() *DeclineStore {
	ds := &DeclineStore{m: make(map[string]map[string]time.Time)}
	go ds.gc()
	return ds
}

func (d *DeclineStore) Add(orderID, driverID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.m[orderID] == nil {
		d.m[orderID] = make(map[string]time.Time)
	}
	d.m[orderID][driverID] = time.Now()
}

func (d *DeclineStore) Get(orderID string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := []string{}
	if m, ok := d.m[orderID]; ok {
		for k := range m {
			out = append(out, k)
		}
	}
	return out
}

func (d *DeclineStore) gc() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		cutoff := time.Now().Add(-30 * time.Minute)
		d.mu.Lock()
		for oid, m := range d.m {
			for did, ts := range m {
				if ts.Before(cutoff) {
					delete(m, did)
				}
			}
			if len(m) == 0 {
				delete(d.m, oid)
			}
		}
		d.mu.Unlock()
	}
}

var Declines *DeclineStore

func init() {
	// Initialize with defaults; will be re-initialized in main with Redis support
	DriverHub = &WsHub{chans: make(map[string]chan []byte)}
	Declines = NewDeclineStore()
}
