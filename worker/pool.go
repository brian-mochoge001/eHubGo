package worker

import (
	"log"
	"sync"
)

type Job interface {
	Execute() error
}

type Pool struct {
	jobQueue   chan Job
	wg         sync.WaitGroup
	maxWorkers int
}

// NewPool initializes a worker pool with the given number of workers.
func NewPool(maxWorkers int, maxQueueSize int) *Pool {
	return &Pool{
		jobQueue:   make(chan Job, maxQueueSize),
		maxWorkers: maxWorkers,
	}
}

// Start launches the background worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	log.Printf("[WorkerPool] Started with %d workers", p.maxWorkers)
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for job := range p.jobQueue {
		// Recover from any panics in jobs to keep the worker alive
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WorkerPool] Worker %d recovered from panic: %v", id, r)
				}
			}()
			
			err := job.Execute()
			if err != nil {
				log.Printf("[WorkerPool] Worker %d job failed: %v", id, err)
			}
		}()
	}
}

// Submit enqueues a job. Returns false if the queue is full.
func (p *Pool) Submit(job Job) bool {
	select {
	case p.jobQueue <- job:
		return true
	default:
		log.Printf("[WorkerPool] Warning: Job queue is full, dropping job")
		return false
	}
}

// Stop gracefully shuts down the worker pool, waiting for pending jobs to complete.
func (p *Pool) Stop() {
	close(p.jobQueue)
	p.wg.Wait()
	log.Printf("[WorkerPool] Stopped")
}
