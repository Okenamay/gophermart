package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// AccrualResponse представляет ответ от системы начислений
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// Client для взаимодействия с системой начислений
type Client struct {
	address        string
	db             *database.Storage
	client         http.Client
	workersCount   int
	rateLimitMutex sync.Mutex
	rateLimitUntil time.Time
}

// NewClient создаёт новый клиент для системы начислений
func NewClient(config *config.Cfg, db *database.Storage) *Client {
	return &Client{
		address:      config.AccrualAddress,
		db:           db,
		client:       http.Client{Timeout: 5 * time.Second},
		workersCount: config.AccrualWorkers,
	}
}

// StartPolling запускает периодический опрос системы начислений
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

// processOrders обрабатывает заказы
func (c *Client) processOrders(ctx context.Context) {
	orders, err := c.db.GetUnprocessedOrders(ctx)
	if err != nil {
		if !errors.Is(err, database.ErrNoOrdersToPoll) {
			logger.Zap.Errorw("Failed to get unprocessed orders", "error", err)
		}
		return
	}

	if len(orders) == 0 {
		return
	}

	logger.Zap.Infow("Found orders to process", "count", len(orders))

	jobs := make(chan string, len(orders))
	results := make(chan error, len(orders))
	var wg sync.WaitGroup

	// Запускаем воркеров
	for w := 1; w <= c.workersCount; w++ {
		wg.Add(1)
		go c.worker(ctx, &wg, jobs, results)
	}

	// Отправляем задания
	for _, order := range orders {
		jobs <- order.Number
	}
	close(jobs)

	// Ожидаем завершения всех воркеров
	wg.Wait()
	close(results)

	// Логируем ошибки, если они были
	for err := range results {
		if err != nil {
			logger.Zap.Errorw("Error processing order in worker", "error", err)
		}
	}
}

// worker представляет единичный исполнитель из пула
func (c *Client) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- error) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case orderNumber, ok := <-jobs:
			if !ok {
				return
			}

			// Проверяем, не активна ли пауза из-за rate limit
			c.rateLimitMutex.Lock()
			pauseUntil := c.rateLimitUntil
			c.rateLimitMutex.Unlock()

			if time.Now().Before(pauseUntil) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(pauseUntil)):
					// Пауза закончилась
				}
			}

			results <- c.updateOrderStatus(ctx, orderNumber)
		}
	}
}

func (c *Client) updateOrderStatus(ctx context.Context, orderNumber string) error {
	url, err := url.JoinPath(c.address, "api", "orders", orderNumber)
	if err != nil {
		logger.Zap.Errorw("Failed to process URL", "error", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Zap.Errorw("Failed to create request", "error", err)
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Zap.Errorw("Failed to query accrual system", "order", orderNumber, "error", err)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			logger.Zap.Errorw("Failed to decode accrual response", "order", orderNumber, "error", err)
			return err
		}
		err := c.db.UpdateOrder(ctx, info.Order, info.Status, info.Accrual)
		if err != nil {
			logger.Zap.Errorw("Failed to update order in DB", "order", orderNumber, "error", err)
			return err
		}

	case http.StatusNoContent:
		logger.Zap.Infow("Order not registered in accrual system yet", "order", orderNumber)
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		logger.Zap.Warnw("Too many requests to accrual system", "order", orderNumber, "retry_after", retryAfter)

		seconds, err := strconv.Atoi(retryAfter)
		if err == nil {
			c.rateLimitMutex.Lock()
			// Устанавливаем паузу, только если текущая пауза не длиннее
			if c.rateLimitUntil.Before(time.Now().Add(time.Duration(seconds) * time.Second)) {
				c.rateLimitUntil = time.Now().Add(time.Duration(seconds) * time.Second)
			}
			c.rateLimitMutex.Unlock()
		}

	case http.StatusInternalServerError:
		logger.Zap.Errorw("Accrual system internal error", "order", orderNumber)
		return err
	default:
		logger.Zap.Errorw("Received unexpected status from accrual system", "order", orderNumber, "status", resp.StatusCode)
		return err
	}
	return nil
}
