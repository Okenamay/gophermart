package config

import (
	"flag"
	"os"
	"strconv"
	"sync"
)

const (
	// Адрес и порт сервера по умолчанию
	RunAddress = ":8080"
	// Адрес подключения к БД по умолчанию
	DatabaseURI = "postgresql://user:password@localhost:5432/gophermart"
	// Адрес системы расчёта начислений по умолчанию
	AccrualAddress = ""
	// Таймаут сервера в секундах
	IdleTimeout = 600
	// Срок истечения действия токена в часах
	TokenExp = 24
	// Ключ для подписи JWT
	AuthKey = "secret_key"
	// Флаг переинициализации БД при старте
	DBReinit = false
)

// Cfg определяет структуру конфигурации
type Cfg struct {
	RunAddress       string
	DatabaseURI      string
	AccrualAddress   string
	IdleTimeout      int
	TokenExpiry      int
	AuthorizationKey string
	DBReinitialize   bool
	MigrateID        string
	MigrateDirection string
}

func parseFlags() *Cfg {
	config := &Cfg{}

	flag.StringVar(&config.RunAddress, "a", RunAddress, "Адрес запуска сервера")
	flag.StringVar(&config.DatabaseURI, "d", DatabaseURI, "Адрес подключения к БД")
	flag.StringVar(&config.AccrualAddress, "r", AccrualAddress, "Адрес системы расчёта начислений")
	flag.BoolVar(&config.DBReinitialize, "dbx", DBReinit, "Реинициализация БД (true/false)")
	flag.StringVar(&config.MigrateID, "migid", "", "ID миграции БД")
	flag.StringVar(&config.MigrateDirection, "migdir", "up", "Направление миграции (up/down)")

	flag.Parse()

	if runAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddress = runAddr
	}
	if dbURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseURI = dbURI
	}
	if accrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.AccrualAddress = accrualAddr
	}
	if dbReinit, ok := os.LookupEnv("DB_REINIT"); ok {
		config.DBReinitialize, _ = strconv.ParseBool(dbReinit)
	}
	if migID, ok := os.LookupEnv("MIGRATION_ID"); ok {
		config.MigrateID = migID
	}
	if migDir, ok := os.LookupEnv("MIGRATION_DIRECTION"); ok {
		config.MigrateDirection = migDir
	}

	// Устанавливаем значения по умолчанию, т.к. флаги и env не определены
	config.IdleTimeout = IdleTimeout
	config.TokenExpiry = TokenExp
	config.AuthorizationKey = AuthKey

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
