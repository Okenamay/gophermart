package router

import (
	"net/http"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/server/handlers"
	mware "github.com/Okenamay/gophermart/internal/server/middleware/auth"
	"github.com/go-chi/chi"
)

// Запуск HTTP-сервера и работа с запросами:
func Launch(conf *config.Cfg) error {
	router := chi.NewRouter()

	router.Use(logger.WithLogging)

	// Эндпоинты для регистрации и логина
	router.Post("/api/user/register", handlers.RegUser(conf))
	router.Post("/api/user/login", handlers.LoginUser(conf))

	router.Group(func(r chi.Router) {
		r.Use(mware.Authenticator(conf))

		// Загрузка номера заказа
		r.Post("/api/user/orders", handlers.AddOrder(conf))
		// Получение списка заказов
		r.Get("/api/user/orders", handlers.ListOrders(conf))
		// Получение баланса
		r.Get("/api/user/balance", handlers.PointsBalance(conf))
		// Списание баллов
		r.Post("/api/user/balance/withdraw", handlers.PointsWithdraw(conf))
		// Получение истории списаний
		r.Get("/api/user/withdrawals", handlers.ListWithdrawals(conf))
	})

	server := http.Server{
		Addr:        conf.RunAddress,
		Handler:     router,
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Second,
	}

	return server.ListenAndServe()
}
