package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// AccrualResponse представляет ответ от системы начислений.
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// Client для взаимодействия с системой начислений.
type Client struct {
	address string
	db      *database.Storage
	client  http.Client
}

// NewClient создаёт новый клиент для системы начислений.
func NewClient(address string, db *database.Storage) *Client {
	return &Client{
		address: address,
		db:      db,
		client:  http.Client{Timeout: 5 * time.Second},
	}
}

// StartPolling запускает периодический опрос системы начислений.
func (c *Client) StartPolling(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	logger.Zap.Info("Starting accrual system polling")

	for {
		select {
		case <-ctx.Done():
			logger.Zap.Info("Stopping accrual polling")
			return
		case <-ticker.C:
			c.processOrders(ctx)
		}
	}
}

func (c *Client) processOrders(ctx context.Context) {
	orders, err := c.db.GetUnprocessedOrders(ctx)
	if err != nil {
		if !errors.Is(err, database.ErrNoOrdersToPoll) {
			logger.Zap.Errorw("Failed to get unprocessed orders", "error", err)
		}
		return
	}

	logger.Zap.Infow("Found orders to process", "count", len(orders))
	for _, order := range orders {
		c.updateOrderStatus(ctx, order.Number)
	}
}

func (c *Client) updateOrderStatus(ctx context.Context, orderNumber string) {
	url := fmt.Sprintf("%s/api/orders/%s", c.address, orderNumber)
	resp, err := c.client.Get(url)
	if err != nil {
		logger.Zap.Errorw("Failed to query accrual system", "order", orderNumber, "error", err)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			logger.Zap.Errorw("Failed to decode accrual response", "order", orderNumber, "error", err)
			return
		}
		err := c.db.UpdateOrder(ctx, info.Order, info.Status, info.Accrual)
		if err != nil {
			logger.Zap.Errorw("Failed to update order in DB", "order", orderNumber, "error", err)
		}
	case http.StatusNoContent:
		logger.Zap.Infow("Order not registered in accrual system yet", "order", orderNumber)
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		logger.Zap.Warnw("Accrual system rate limit exceeded", "retry_after", retryAfter)
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	default:
		logger.Zap.Errorw("Received unexpected status from accrual system", "order", orderNumber, "status", resp.StatusCode)
	}
}
