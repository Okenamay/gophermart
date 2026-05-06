package handlers

import "time"

// UserCredentials используется для регистрации и логина
type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Order описывает структуру заказа для ответа API
type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance описывает баланс пользователя
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalRequest используется для запроса на списание
type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Withdrawal описывает операцию списания для ответа API
type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
