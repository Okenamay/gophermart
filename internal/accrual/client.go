package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	// "github.com/Okenamay/gophermart/internal/storage" // Вам понадобится ваш пакет storage
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
	// db      storage.Storage // Интерфейс вашего хранилища
	client http.Client
}

// NewClient создаёт новый клиент для системы начислений.
func NewClient(address string /*, db storage.Storage*/) *Client {
	return &Client{
		address: address,
		// db:      db,
		client: http.Client{Timeout: 5 * time.Second},
	}
}

// StartPolling запускает периодический опрос системы начислений.
func (c *Client) StartPolling(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Опрашиваем каждые 5 секунд
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
	// TODO: Замените на реальный вызов к вашей БД
	// orders, err := c.db.GetUnprocessedOrders(ctx)
	// if err != nil {
	// 	logger.Zap.Errorf("Failed to get unprocessed orders: %v", err)
	// 	return
	// }
	//
	// if len(orders) == 0 {
	// 	return
	// }
	//
	// logger.Zap.Info("Found orders to process", "count", len(orders))
	// for _, order := range orders {
	// 	c.updateOrderStatus(ctx, order.Number)
	// }

	// Mock-данные для демонстрации
	mockOrders := []string{"12345678903", "9278923470"}
	logger.Zap.Infof("Found %d orders to process", len(mockOrders))
	for _, orderNum := range mockOrders {
		c.updateOrderStatus(ctx, orderNum)
	}
}

func (c *Client) updateOrderStatus(ctx context.Context, orderNumber string) {
	url := fmt.Sprintf("%s/api/orders/%s", c.address, orderNumber)
	resp, err := c.client.Get(url)
	if err != nil {
		logger.Zap.Error("Failed to query accrual system", "order", orderNumber, "error", err)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualInfo AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualInfo); err != nil {
			logger.Zap.Errorf("Failed to decode accrual response for order %s: %v", orderNumber, err)
			return
		}
		logger.Zap.Info("Received accrual update", "order", orderNumber, "status", accrualInfo.Status, "accrual", accrualInfo.Accrual)
		// TODO: Обновите заказ в вашей БД
		// err := c.db.UpdateOrder(ctx, accrualInfo.Order, accrualInfo.Status, accrualInfo.Accrual)
		// if err != nil {
		// 	logger.Zap.Errorf("Failed to update order %s in DB: %v", orderNumber, err)
		// }

	case http.StatusNoContent:
		logger.Zap.Warnf("Order %s not registered in accrual system yet", orderNumber)

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		logger.Zap.Warn("Accrual system rate limit exceeded", "retry_after", retryAfter)
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			time.Sleep(time.Duration(seconds) * time.Second)
		}

	default:
		logger.Zap.Errorf("Received unexpected status %d from accrual system for order %s", resp.StatusCode, orderNumber)
	}
}
