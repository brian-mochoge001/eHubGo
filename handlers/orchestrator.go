package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"ehubgo/db"
	"github.com/google/uuid"
)

type OrderCoordinator struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewOrderCoordinator(queries *db.Queries, dbConn *sql.DB) *OrderCoordinator {
	return &OrderCoordinator{
		Queries: queries,
		DB:      dbConn,
	}
}

// OrchestrateCheckout handles splitting cart items into sub-orders atomically.
func (oc *OrderCoordinator) OrchestrateCheckout(ctx context.Context, userID string, addressID string, cartItems []db.GetCartItemsByUserIDRow) (string, error) {
	if len(cartItems) == 0 {
		return "", fmt.Errorf("cart is empty")
	}

	tx, err := oc.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	qtx := oc.Queries.WithTx(tx)

	// 1. Create Parent Order
	parentOrderID := uuid.New().String()
	totalAmount := 0.0
	for _, item := range cartItems {
		price, _ := strconv.ParseFloat(item.Price, 64)
		totalAmount += price * float64(item.Quantity)
	}

	_, err = qtx.CreateOrder(ctx, db.CreateOrderParams{
		ID:                parentOrderID,
		UserID:            userID,
		TotalAmount:       fmt.Sprintf("%.2f", totalAmount),
		Currency:          "Ksh",
		Status:            "pending",
		ShippingAddressID: sql.NullString{String: addressID, Valid: true},
	})
	if err != nil {
		return "", err
	}

	// 2. Group items by business
	groupedItems := make(map[string][]db.GetCartItemsByUserIDRow)
	for _, item := range cartItems {
		groupedItems[item.BusinessID] = append(groupedItems[item.BusinessID], item)
	}

	// 3. Create Sub-Orders
	for businessID, items := range groupedItems {
		subOrderID := uuid.New().String()
		subTotal := 0.0
		for _, item := range items {
			price, _ := strconv.ParseFloat(item.Price, 64)
			subTotal += price * float64(item.Quantity)
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO orders (id, user_id, parent_order_id, total_amount, status, shipping_address_id)
			VALUES ($1, $2, $3, $4, 'pending', $5)
		`, subOrderID, userID, parentOrderID, fmt.Sprintf("%.2f", subTotal), addressID)
		if err != nil {
			return "", err
		}

		// 4. Create Order Items and Lock/Decrement Stock
		for _, item := range items {
			err = qtx.LockAndDecrementStock(ctx, db.LockAndDecrementStockParams{
				StockQuantity: item.Quantity,
				ID:            item.ItemID,
			})
			if err != nil {
				return "", fmt.Errorf("stock reservation failed for item %s: %w", item.ItemID, err)
			}

			_, err = qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
				ID:         uuid.New().String(),
				OrderID:    subOrderID,
				BusinessID: businessID,
				ItemID:     item.ItemID,
				ItemType:   item.ItemType,
				Quantity:   item.Quantity,
				UnitPrice:  item.Price,
			})
			if err != nil {
				return "", err
			}
		}
	}

	// 5. Clear Cart
	err = qtx.ClearCart(ctx, userID)
	if err != nil {
		return "", err
	}

	return parentOrderID, tx.Commit()
}
