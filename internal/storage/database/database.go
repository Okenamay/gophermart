package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/storage/migration"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage инкапсулирует пул соединений с БД
type Storage struct {
	DBPool *pgxpool.Pool
}

// New инициализирует подключение к базе данных, выполняет реинициализацию или миграции
func New(ctx context.Context, conf *config.Cfg) (*Storage, error) {
	logger.Zap.Info("Initializing database connection")
	pool, err := pgxpool.New(ctx, conf.DatabaseURI)
	if err != nil {
		logger.Zap.Errorw("Failed to connect to database", "error", err)
		return nil, err
	}
	storage := &Storage{DBPool: pool}
	if err := storage.Ping(ctx); err != nil {
		return nil, err
	}
	logger.Zap.Info("Database connection established successfully")

	// Запускаем миграции при старте приложения
	if err := migration.MigrateLauncher(ctx, conf); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return storage, nil
}

// Ping проверяет доступность базы данных
func (s *Storage) Ping(ctx context.Context) error {
	logger.Zap.Info("Pinging database")
	if s.DBPool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := s.DBPool.Ping(ctx); err != nil {
		logger.Zap.Errorw("Database ping failed", "error", err)
		return err
	}
	logger.Zap.Info("Database ping successful")
	return nil
}

// Close закрывает пул соединений с базой данных
func (s *Storage) Close() {
	if s.DBPool != nil {
		s.DBPool.Close()
		logger.Zap.Info("Database connection pool closed.")
	}
}
