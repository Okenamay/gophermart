package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Okenamay/gophermart/internal/accrual"
	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/server/router"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

func main() {
	appLogger, err := logger.InitLogger()
	if err != nil {
		// Если логгер не стартовал, мы не можем даже это залогировать. Паникуем.
		panic("failed to initialize logger: " + err.Error())
	}
	defer appLogger.Sync()

	conf := config.InitConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Инициализация хранилища
	db, err := database.New(ctx, conf, appLogger)
	if err != nil {
		appLogger.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	if conf.AccrualAddress != "" {
		// Передаём зависимость от БД и конфига в клиент системы начислений
		accrualClient := accrual.NewClient(conf, db, appLogger)
		go accrualClient.StartPolling(ctx)
	} else {
		appLogger.Warn("ACCRUAL_SYSTEM_ADDRESS is not set. Accrual polling is disabled.")
	}

	appLogger.Infow("Starting server", "address", conf.RunAddress)

	// Передаем логгер в роутер
	r := router.SetupRouter(conf, db, appLogger)
	server := &http.Server{
		Addr:        conf.RunAddress,
		Handler:     r,
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		appLogger.Infow("Starting server", "address", conf.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Ожидаем сигнала завершения для graceful shutdown
	<-ctx.Done()

	appLogger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatalw("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exited properly")

}
