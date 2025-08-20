package main

import (
	"context"

	"github.com/Okenamay/gophermart/internal/accrual"
	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/server/router"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

var err error

func main() {
	if err = logger.InitLogger(); err != nil {
		logger.Zap.Fatalw("Failed to start logger", "error", err)
	}

	conf := config.InitConfig()
	ctx := context.Background()

	// Инициализация хранилища из нового пакета
	db, err := database.New(ctx, conf)
	if err != nil {
		logger.Zap.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Проверяем, задан ли адрес системы начислений
	if conf.AccrualAddress != "" {
		// Инициализируем и запускаем клиент системы начислений
		accrualClient := accrual.NewClient(conf.AccrualAddress /*, db*/)
		go accrualClient.StartPolling(ctx)
	} else {
		logger.Zap.Warn("ACCRUAL_SYSTEM_ADDRESS is not set. Accrual polling is disabled.")
	}

	logger.Zap.Info("Starting server", "address", conf.RunAddress)

	err = router.Launch(conf)
	if err != nil {
		logger.Zap.Fatalw("Failed to start server", "error", err)
	}

	defer logger.Zap.Sync()
}
