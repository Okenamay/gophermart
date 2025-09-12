package config

import (
	"flag"
	"os"
	"sync"
)

const (
	// Адрес и порт сервера по умолчанию
	runAddress = ":8080"
	// Адрес подключения к БД по умолчанию
	databaseURI = "postgresql://user:password@localhost:5432/gophermart"
	// Адрес системы расчёта начислений по умолчанию
	accrualAddress = ""
	// Таймаут сервера в секундах
	idleTimeout = 600
	// Срок истечения действия токена в часах
	tokenExp = 24
	// Ключ для подписи JWT
	authKey = "secret_key"
)

// Cfg определяет структуру конфигурации
type Cfg struct {
	RunAddress       string
	DatabaseURI      string
	AccrualAddress   string
	IdleTimeout      int
	TokenExpiry      int
	AuthorizationKey string
}

func parseFlags() *Cfg {
	config := &Cfg{}

	// Инициализируцемся дефолтными значениями:
	config.RunAddress = runAddress
	config.DatabaseURI = databaseURI
	config.AccrualAddress = accrualAddress
	config.IdleTimeout = idleTimeout
	config.TokenExpiry = tokenExp
	config.AuthorizationKey = authKey

	// Переписываем дефолтные env'ами:
	if runAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddress = runAddr
	}
	if dbURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseURI = dbURI
	}
	if accrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.AccrualAddress = accrualAddr
	}

	// Преписываем всё флагами:
	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "Адрес запуска сервера")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "Адрес подключения к БД")
	flag.StringVar(&config.AccrualAddress, "r", config.AccrualAddress, "Адрес системы расчёта начислений")

	flag.Parse()

	// Устанавливаем значения по умолчанию, т.к. флаги и env не определены
	config.IdleTimeout = idleTimeout
	config.TokenExpiry = tokenExp
	config.AuthorizationKey = authKey

	return config
}

// InitConfig инициализирует конфигурацию синглтоном
func InitConfig() *Cfg {
	var (
		once   sync.Once
		config *Cfg
	)
	once.Do(func() {
		config = parseFlags()
	})
	return config
}
