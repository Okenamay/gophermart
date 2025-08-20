package migration

import "github.com/Okenamay/gophermart/internal/config"

// MigrationEntry описывает миграцию
type MigrationEntry struct {
	ID      string
	UpSQL   string
	DownSQL string
}

// migrations содержит все доступные миграции
var migrations = map[string]MigrationEntry{
	"20250820000000": {
		ID: "20250820000000",
		UpSQL: `
            CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

            CREATE TABLE IF NOT EXISTS users (
                id SERIAL PRIMARY KEY,
                login VARCHAR(255) UNIQUE NOT NULL,
                password_hash VARCHAR(255) NOT NULL,
                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            );

            CREATE TABLE IF NOT EXISTS orders (
                id SERIAL PRIMARY KEY,
                user_id INTEGER REFERENCES users(id),
                number VARCHAR(255) UNIQUE NOT NULL,
                status order_status NOT NULL DEFAULT 'NEW',
                accrual NUMERIC(10, 2),
                uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            );

            CREATE TABLE IF NOT EXISTS withdrawals (
                id SERIAL PRIMARY KEY,
                user_id INTEGER REFERENCES users(id),
                order_number VARCHAR(255) NOT NULL,
                sum NUMERIC(10, 2) NOT NULL,
                processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            );
        `,
		DownSQL: `
            DROP TABLE IF EXISTS withdrawals;
            DROP TABLE IF EXISTS orders;
            DROP TABLE IF EXISTS users;
            DROP TYPE IF EXISTS order_status;
        `,
	},
}

// DeliverMigration находит и возвращает миграцию по её ID
func DeliverMigration(conf *config.Cfg) MigrationEntry {
	return migrations[conf.MigrateID]
}
