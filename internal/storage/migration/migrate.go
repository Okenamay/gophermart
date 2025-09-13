package migration

import (
	"context"
	"database/sql"
	"embed"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// MigrateLauncher запускает процесс миграции
func MigrateLauncher(ctx context.Context, conf *config.Cfg) error {

	db, err := sql.Open("pgx", conf.DatabaseURI)
	if err != nil {
		logger.Zap.Errorw("Failed to open DB connection for migration", "error", err)
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		logger.Zap.Errorw("Failed to ping DB for migration", "error", err)
		return err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Zap.Errorw("Failed to set goose dialect", "error", err)
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		logger.Zap.Errorw("Failed to apply migrations", "error", err)
		return err
	}

	return nil
}
