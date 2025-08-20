package migration

import (
	"context"
	"fmt"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrateLauncher запускает процесс миграции или отката.
func MigrateLauncher(ctx context.Context, dbpool *pgxpool.Pool, conf *config.Cfg) error {
	if conf.MigrateID == "" {
		logger.Zap.Info("Migration ID is not provided. Skipping.")
		return nil
	}

	logger.Zap.Infow("Attempting to run migration",
		"id", conf.MigrateID,
		"direction", conf.MigrateDirection,
	)

	migration := DeliverMigration(conf)
	if migration.ID == "" {
		return fmt.Errorf("unknown migration ID: %s", conf.MigrateID)
	}

	var err error
	var sql string

	switch conf.MigrateDirection {
	case "up":
		sql = migration.UpSQL
		_, err = dbpool.Exec(ctx, sql)
		if err != nil {
			logger.Zap.Errorw("Failed to apply migration", "id", migration.ID, "error", err)
			return err
		}
		logger.Zap.Infow("Successfully applied migration", "id", migration.ID)
	case "down":
		sql = migration.DownSQL
		_, err = dbpool.Exec(ctx, sql)
		if err != nil {
			logger.Zap.Errorw("Failed to apply rollback", "id", migration.ID, "error", err)
			return err
		}
		logger.Zap.Infow("Successfully applied rollback", "id", migration.ID)
	default:
		return fmt.Errorf("incorrect migration direction: %s", conf.MigrateDirection)
	}

	return nil
}
