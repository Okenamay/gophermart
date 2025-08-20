package router

import (
	"net/http"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/server/handlers"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/go-chi/chi/v5"
)

// Launch запускает HTTP-сервер и настраивает маршруты
func Launch(conf *config.Cfg, db *database.Storage) error {
	r := chi.NewRouter()

	// Подключаем middleware для логирования
	r.Use(logger.WithLogging)

	// Публичные маршруты
	r.Group(func(r chi.Router) {
		h := handlers.New(conf, db)
		r.Post("/api/user/register", h.RegisterUser)
		r.Post("/api/user/login", h.LoginUser)
	})

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(conf))
		h := handlers.New(conf, db)
		r.Post("/api/user/orders", h.AddOrder)
		r.Get("/api/user/orders", h.ListOrders)
		r.Get("/api/user/balance", h.PointsBalance)
		r.Post("/api/user/balance/withdraw", h.PointsWithdraw)
		r.Get("/api/user/withdrawals", h.ListWithdrawals)
	})

	server := http.Server{
		Addr:        conf.RunAddress,
		Handler:     r,
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Second,
	}

	return server.ListenAndServe()
}
