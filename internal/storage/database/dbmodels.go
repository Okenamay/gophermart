package database

import (
	"errors"
	"time"
)

// --- МОДЕЛИ ---

// User представляет модель пользователя в базе данных.
type User struct {
	ID           int
	Login        string
	PasswordHash string
}

// Order представляет модель заказа в базе данных.
type Order struct {
	ID         int
	UserID     int
	Number     string
	Status     string
	Accrual    *float64 // Используем указатель, чтобы корректно обрабатывать NULL
	UploadedAt time.Time
}

// --- ОШИБКИ ---

var (
	ErrLoginConflict  = errors.New("login already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrOrderConflict  = errors.New("order number already exists")
	ErrOrderNotFound  = errors.New("order not found")
	ErrNoOrdersFound  = errors.New("no orders found for this user")
	ErrNoOrdersToPoll = errors.New("no orders to poll")
)
