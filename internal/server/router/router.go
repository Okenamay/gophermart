package router

import (
	"net/http"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/server/handlers"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Launch запускает HTTP-сервер и настраивает маршруты, принимает логгер
func SetupRouter(conf *config.Cfg, db *database.Storage, appLogger *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	// Подключаем middleware для логирования
	r.Use(logger.WithLogging(appLogger))

	// Создаем экземпляр хендлера
	h := handlers.New(conf, db, appLogger)

	// Публичные маршруты
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.RegisterUser)
		r.Post("/api/user/login", h.LoginUser)
	})

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(conf, appLogger))
		r.Post("/api/user/orders", h.AddOrder)
		r.Get("/api/user/orders", h.ListOrders)
		r.Get("/api/user/balance", h.PointsBalance)
		r.Post("/api/user/balance/withdraw", h.PointsWithdraw)
		r.Get("/api/user/withdrawals", h.ListWithdrawals)
	})

	return r
}
