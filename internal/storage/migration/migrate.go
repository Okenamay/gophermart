package migration

import (
	"context"
	"database/sql"
	"embed"

	"github.com/Okenamay/gophermart/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// MigrateLauncher запускает процесс миграции
func MigrateLauncher(ctx context.Context, conf *config.Cfg, appLogger *zap.SugaredLogger) error {
	db, err := sql.Open("pgx", conf.DatabaseURI)
	if err != nil {
		appLogger.Errorw("Failed to open DB connection for migration", "error", err)
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		appLogger.Errorw("Failed to ping DB for migration", "error", err)
		return err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		appLogger.Errorw("Failed to set goose dialect", "error", err)
		return err
	}

	appLogger.Info("Running database migrations...")

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		appLogger.Errorw("Failed to apply migrations", "error", err)
		return err
	}

	appLogger.Info("Database migrations applied successfully")
	return nil
}
