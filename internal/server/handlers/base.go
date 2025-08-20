package handlers

import (
	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// Handler - структура для хранения зависимостей обработчиков.
type Handler struct {
	Config *config.Cfg
	DB     *database.Storage
}

// New создает новый экземпляр Handler.
func New(cfg *config.Cfg, db *database.Storage) *Handler {
	return &Handler{
		Config: cfg,
		DB:     db,
	}
}
