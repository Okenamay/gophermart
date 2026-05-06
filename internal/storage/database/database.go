package database

import (
	"context"

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
	storage := &Storage{
		DBPool:    pool,
		appLogger: appLogger,
	}
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
	return s.DBPool.Ping(ctx)
}

// Close закрывает пул соединений с базой данных
func (s *Storage) Close() {
	s.appLogger.Info("Closing database connection pool")
	s.DBPool.Close()
}
