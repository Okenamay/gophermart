package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/storage/migration"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Storage инкапсулирует пул соединений с БД
type Storage struct {
	DBPool    *pgxpool.Pool
	appLogger *zap.SugaredLogger
}

// New инициализирует подключение к базе данных, выполняет реинициализацию или миграции
func New(ctx context.Context, conf *config.Cfg, appLogger *zap.SugaredLogger) (*Storage, error) {
	appLogger.Info("Initializing database connection")
	pool, err := pgxpool.New(ctx, conf.DatabaseURI)
	if err != nil {
		appLogger.Errorw("Failed to connect to database", "error", err)
		return nil, err
	}
	storage := &Storage{DBPool: pool}
	if err := storage.Ping(ctx); err != nil {
		return nil, err
	}
	appLogger.Info("Database connection established successfully")

	// Запускаем миграции при старте приложения
	if err := migration.MigrateLauncher(ctx, conf, appLogger); err != nil {
		appLogger.Errorw("Failed to apply migrations", "error", err)
		return nil, err
	}

	return storage, nil
}

// Ping проверяет доступность базы данных
func (s *Storage) Ping(ctx context.Context) error {
	s.appLogger.Info("Pinging database")
	if s.DBPool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := s.DBPool.Ping(ctx); err != nil {
		s.appLogger.Errorw("Database ping failed", "error", err)
		return err
	}
	s.appLogger.Info("Database ping successful")
	return nil
}

// Close закрывает пул соединений с базой данных
func (s *Storage) Close() {
	if s.DBPool != nil {
		s.DBPool.Close()
		s.appLogger.Info("Database connection pool closed.")
	}
}
