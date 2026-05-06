package database

import (
	"errors"
	"time"
)

// User представляет модель пользователя в базе данных
type User struct {
	ID           int
	Login        string
	PasswordHash string
}

// Order представляет модель заказа в базе данных
type Order struct {
	ID         int
	UserID     int
	Number     string
	Status     string
	Accrual    *float64 // Используем указатель, чтобы корректно обрабатывать NULL
	UploadedAt time.Time
}

// Withdrawal представляет модель списания в базе данных
type Withdrawal struct {
	ID          int
	UserID      int
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

// Balance представляет баланс пользователя, рассчитанный из БД
type Balance struct {
	Current   float64
	Withdrawn float64
}

// Кастомные ошибки
var (
	ErrLoginConflict      = errors.New("login already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrOrderConflict      = errors.New("order number already exists")
	ErrOrderNotFound      = errors.New("order not found")
	ErrNoOrdersFound      = errors.New("no orders found for this user")
	ErrNoOrdersToPoll     = errors.New("no orders to poll")
	ErrInsufficientFunds  = errors.New("insufficient funds on balance")
	ErrNoWithdrawalsFound = errors.New("no withdrawals found for this user")
)
