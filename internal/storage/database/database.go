package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/storage/migration"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage инкапсулирует пул соединений с БД.
type Storage struct {
	DBPool *pgxpool.Pool
}

// New инициализирует подключение к базе данных, выполняет реинициализацию или миграции.
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

	// Проверяем, существует ли схема, и создаем ее, если необходимо.
	schemaExists, err := storage.checkSchema(ctx)
	if err != nil {
		return nil, err
	}

	if conf.DBReinitialize || !schemaExists {
		if !schemaExists {
			logger.Zap.Warn("Database schema not found. Forcing re-initialization.")
		} else {
			logger.Zap.Warn("DBReinitialize flag is set. Re-creating database schema.")
		}
		if err := storage.reinitialize(ctx); err != nil {
			return nil, err
		}
	}

	if conf.MigrateID != "" {
		if err := migration.MigrateLauncher(ctx, storage.DBPool, conf); err != nil {
			return nil, err
		}
	}
	return storage, nil
}

// checkSchema проверяет наличие основной таблицы 'users' в БД.
func (s *Storage) checkSchema(ctx context.Context) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = 'users');"
	err := s.DBPool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		logger.Zap.Errorw("Failed to check database schema", "error", err)
		return false, err
	}
	return exists, nil
}

// --- ФУНКЦИИ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ ---

// CreateUser создает нового пользователя в базе данных.
func (s *Storage) CreateUser(ctx context.Context, login, passwordHash string) (int, error) {
	var userID int
	err := s.DBPool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2)
         ON CONFLICT (login) DO NOTHING
         RETURNING id`,
		login, passwordHash).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Zap.Warnw("Login conflict on user creation", "login", login)
			return 0, ErrLoginConflict
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrLoginConflict
		}
		logger.Zap.Errorw("Failed to create user", "login", login, "error", err)
		return 0, err
	}
	logger.Zap.Infow("Successfully created user", "login", login, "userID", userID)
	return userID, nil
}

// GetUserByLogin находит пользователя по логину.
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	user := &User{Login: login}
	err := s.DBPool.QueryRow(ctx,
		"SELECT id, password_hash FROM users WHERE login = $1",
		login).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		logger.Zap.Errorw("Failed to get user by login", "login", login, "error", err)
		return nil, err
	}
	return user, nil
}

// --- ФУНКЦИИ ДЛЯ РАБОТЫ С ЗАКАЗАМИ ---

// CreateOrder сохраняет новый заказ в БД.
func (s *Storage) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	_, err := s.DBPool.Exec(ctx,
		"INSERT INTO orders (user_id, number) VALUES ($1, $2)",
		userID, orderNumber)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrOrderConflict
		}
		logger.Zap.Errorw("Failed to create order", "orderNumber", orderNumber, "error", err)
		return err
	}
	return nil
}

// GetOrderByNumber получает заказ по его номеру.
func (s *Storage) GetOrderByNumber(ctx context.Context, number string) (*Order, error) {
	order := &Order{Number: number}
	err := s.DBPool.QueryRow(ctx,
		"SELECT id, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1",
		number).Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		logger.Zap.Errorw("Failed to get order by number", "number", number, "error", err)
		return nil, err
	}
	return order, nil
}

// GetOrdersByUser получает все заказы пользователя, отсортированные по времени загрузки.
func (s *Storage) GetOrdersByUser(ctx context.Context, userID int) ([]Order, error) {
	rows, err := s.DBPool.Query(ctx,
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC",
		userID)
	if err != nil {
		logger.Zap.Errorw("Failed to query user orders", "userID", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			logger.Zap.Errorw("Failed to scan order row", "userID", userID, "error", err)
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		logger.Zap.Errorw("Error iterating user orders rows", "userID", userID, "error", err)
		return nil, err
	}

	if len(orders) == 0 {
		return nil, ErrNoOrdersFound
	}

	return orders, nil
}

// GetUnprocessedOrders получает заказы со статусами NEW или PROCESSING.
func (s *Storage) GetUnprocessedOrders(ctx context.Context) ([]Order, error) {
	rows, err := s.DBPool.Query(ctx,
		"SELECT id, number FROM orders WHERE status IN ('NEW', 'PROCESSING')")
	if err != nil {
		logger.Zap.Errorw("Failed to query unprocessed orders", "error", err)
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number); err != nil {
			logger.Zap.Errorw("Failed to scan unprocessed order row", "error", err)
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		logger.Zap.Errorw("Error iterating unprocessed orders rows", "error", err)
		return nil, err
	}

	if len(orders) == 0 {
		return nil, ErrNoOrdersToPoll
	}

	return orders, nil
}

// UpdateOrder обновляет статус и сумму начисления для заказа.
func (s *Storage) UpdateOrder(ctx context.Context, number, status string, accrual float64) error {
	res, err := s.DBPool.Exec(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		status, accrual, number)
	if err != nil {
		logger.Zap.Errorw("Failed to update order", "number", number, "error", err)
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// --- ФУНКЦИИ ДЛЯ РАБОТЫ С БАЛАНСОМ И СПИСАНИЯМИ ---

// GetUserBalance рассчитывает текущий и списанный баланс пользователя.
func (s *Storage) GetUserBalance(ctx context.Context, userID int) (*Balance, error) {
	var totalAccrual float64
	err := s.DBPool.QueryRow(ctx,
		"SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'",
		userID).Scan(&totalAccrual)
	if err != nil {
		logger.Zap.Errorw("Failed to calculate total accrual", "userID", userID, "error", err)
		return nil, err
	}

	var totalWithdrawn float64
	err = s.DBPool.QueryRow(ctx,
		"SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1",
		userID).Scan(&totalWithdrawn)
	if err != nil {
		logger.Zap.Errorw("Failed to calculate total withdrawn", "userID", userID, "error", err)
		return nil, err
	}

	return &Balance{
		Current:   totalAccrual - totalWithdrawn,
		Withdrawn: totalWithdrawn,
	}, nil
}

// CreateWithdrawal создает запись о списании в транзакции, проверяя баланс.
func (s *Storage) CreateWithdrawal(ctx context.Context, userID int, orderNumber string, sum float64) error {
	tx, err := s.DBPool.Begin(ctx)
	if err != nil {
		logger.Zap.Errorw("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	balance, err := s.getUserBalanceInTx(ctx, tx, userID)
	if err != nil {
		// Логирование происходит внутри getUserBalanceInTx
		return err
	}

	if balance.Current < sum {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)",
		userID, orderNumber, sum)
	if err != nil {
		logger.Zap.Errorw("Failed to insert withdrawal in transaction", "userID", userID, "error", err)
		return err
	}

	return tx.Commit(ctx)
}

// getUserBalanceInTx - вспомогательная функция для получения баланса внутри транзакции.
func (s *Storage) getUserBalanceInTx(ctx context.Context, tx pgx.Tx, userID int) (*Balance, error) {
	var totalAccrual, totalWithdrawn float64
	err := tx.QueryRow(ctx, "SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'", userID).Scan(&totalAccrual)
	if err != nil {
		logger.Zap.Errorw("Failed to calculate total accrual in transaction", "userID", userID, "error", err)
		return nil, err
	}
	err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1", userID).Scan(&totalWithdrawn)
	if err != nil {
		logger.Zap.Errorw("Failed to calculate total withdrawn in transaction", "userID", userID, "error", err)
		return nil, err
	}
	return &Balance{Current: totalAccrual - totalWithdrawn, Withdrawn: totalWithdrawn}, nil
}

// GetWithdrawalsByUser получает историю списаний пользователя.
func (s *Storage) GetWithdrawalsByUser(ctx context.Context, userID int) ([]Withdrawal, error) {
	rows, err := s.DBPool.Query(ctx,
		"SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at ASC",
		userID)
	if err != nil {
		logger.Zap.Errorw("Failed to query user withdrawals", "userID", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var withdrawals []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			logger.Zap.Errorw("Failed to scan withdrawal row", "userID", userID, "error", err)
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		logger.Zap.Errorw("Error iterating withdrawals rows", "userID", userID, "error", err)
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, ErrNoWithdrawalsFound
	}

	return withdrawals, nil
}

// --- СЛУЖЕБНЫЕ ФУНКЦИИ БД ---

// reinitialize полностью удаляет и воссоздаёт структуру таблиц в БД.
func (s *Storage) reinitialize(ctx context.Context) error {
	logger.Zap.Info("Starting database re-initialization")
	sql := `
        DROP TABLE IF EXISTS withdrawals; DROP TABLE IF EXISTS orders; DROP TABLE IF EXISTS users; DROP TYPE IF EXISTS order_status;
        CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
        CREATE TABLE users (id SERIAL PRIMARY KEY, login VARCHAR(255) UNIQUE NOT NULL, password_hash VARCHAR(255) NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
        CREATE TABLE orders (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id), number VARCHAR(255) UNIQUE NOT NULL, status order_status NOT NULL DEFAULT 'NEW', accrual NUMERIC(10, 2), uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
        CREATE TABLE withdrawals (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id), order_number VARCHAR(255) NOT NULL, sum NUMERIC(10, 2) NOT NULL, processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
    `
	_, err := s.DBPool.Exec(ctx, sql)
	if err != nil {
		logger.Zap.Errorw("Failed to re-initialize database", "error", err)
		return err
	}
	logger.Zap.Info("Database re-initialized successfully")
	return nil
}

// Ping проверяет доступность базы данных.
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

// Close закрывает пул соединений с базой данных.
func (s *Storage) Close() {
	if s.DBPool != nil {
		s.DBPool.Close()
		logger.Zap.Info("Database connection pool closed.")
	}
}
