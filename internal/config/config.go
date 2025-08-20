package config

import (
	"flag"
	"os"
	"strconv"
	"sync"

	logger "github.com/Okenamay/gophermart/internal/logger/zap"
)

// Дефолтные значения до применения флагов:
const (
	// Обязательные по ТЗ:
	RunAddress     = ":8080"                                        // Адрес и порт сервера
	DatabaseURI    = "postgresql://tester:1234@localhost:5432/pgdb" // DSN по умолчанию
	AccrualAddress = ""                                             // Адрес системы расчёта начислений.

	// Нужные для работы:
	IdleTimeout = 600 // Таймаут сервера в секундах
	TokenExp    = 24  // Срок истечения действия токена.

	// Полезные:
	MigrID   = ""           // Заглушка
	MigrDir  = ""           // Заглушка
	DBReinit = true         // Флаг переинициализации БД при старте
	AuthKey  = "secret_key" // Ключ авторизации.

	// Грязное легаси:
	// ShortIDLen  = 10                                             // Длина короткого идентификатора
	// ShortIDAddr = "http://localhost:8080"                        // Адрес и порт для коротких ID
	// SaveFile    = "/tmp/short-url-db.json"                       // Имя файла-хранилища
	// Verbose     = false                                          // Флаг детальности логов. !!! Временная заглушка
)

type Cfg struct {
	// Обязательные по ТЗ:
	RunAddress     string
	DatabaseURI    string
	AccrualAddress string

	// Нужные для работы:
	IdleTimeout int
	TokenExpiry int

	// Полезные:
	MigrateID        string
	MigrateDirection string
	DBReinitialize   bool
	AuthorizationKey string

	// Грязное легаси:
	// ShortIDLen        int
	// ShortIDServerPort string
	// SaveFilePath string
	// MemMode     string
	// LogVerbose        bool
}

func parseFlags() *Cfg {
	config := &Cfg{}

	// Тут обязательные флаги по ТЗ:
	flag.StringVar(&config.RunAddress, "a", RunAddress,
		"Адрес запуска сервера в формате host:port или :port")
	flag.StringVar(&config.DatabaseURI, "d", "",
		"DSN подключения к СУБД PostgreSQL")
	flag.StringVar(&config.AccrualAddress, "r", AccrualAddress,
		"Адрес расположения системы расчёта начислений")

	// Тут флаги, которые нужны для работы:
	flag.IntVar(&config.IdleTimeout, "t", IdleTimeout,
		"Таймаут сервера – целое число, желательно от 10 до 600")
	flag.IntVar(&config.TokenExpiry, "txp", TokenExp,
		"Срок истечения токена, часов")

	// Тут полезные флаги:
	flag.StringVar(&config.MigrateID, "migid", MigrID,
		"ID миграции БД в формате YYYYMMDDHHMMSS")
	flag.StringVar(&config.MigrateDirection, "migdir", MigrDir,
		"Направление миграции БД (up = миграция, down = роллбек)")
	flag.BoolVar(&config.DBReinitialize, "dbx", DBReinit,
		"Реинициализация БД (bool)")
	flag.StringVar(&config.AuthorizationKey, "k", AuthKey,
		"Ключ для генерации JWT-токена")

	// А тут легаси:
	// flag.IntVar(&config.ShortIDLen, "l", ShortIDLen,
	// 	"Длина короткого ID – целое число от 8 до 32")
	// flag.StringVar(&config.ShortIDServerPort, "b", ShortIDAddr,
	// 	"Адрес коротких ID в формате host:port/path")
	// flag.StringVar(&config.SaveFilePath, "f", "",
	// 	"Адрес места хранения файла")
	// flag.BoolVar(&config.LogVerbose, "log", Verbose,
	// 	"Вывод подробного лога (bool)")
	flag.Parse()

	var databaseURI string

	// Тут переменные среды обязательные по ТЗ:
	if runAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok && runAddr != "" {
		config.RunAddress = runAddr
	}

	if databaseURI, ok := os.LookupEnv("DATABASE_URI"); ok && databaseURI != "" {
		config.DatabaseURI = databaseURI
		logger.Zap.Info("Loaded database URI from env", "db_uri", databaseURI)
	}

	if accrualAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok && accrualAddress != "" {
		config.AccrualAddress = accrualAddress
		logger.Zap.Info("Loaded accrual system address from env", "accrual_address", accrualAddress)
	}

	// Тут нужные для работы переменные среды:
	if tokenExpiryStr, ok := os.LookupEnv("TOKEN_EXPIRY"); ok && tokenExpiryStr != "" {
		tokenExpiry, err := strconv.Atoi(tokenExpiryStr)
		if err == nil {
			config.TokenExpiry = tokenExpiry
			logger.Zap.Info("Loaded token expiry time from env", "token_expiry", tokenExpiryStr)
		} else {
			logger.Zap.Error("Could not process TOKEN_EXPIRY", "error", err)
		}
	}

	// Тут полезные переменные среды:
	if migrateID, ok := os.LookupEnv("MIGRATION_ID"); ok && migrateID != "" {
		config.MigrateID = migrateID
		logger.Zap.Info("Loaded DB migration ID from env", "migr_dir", migrateID)
	}

	if migrateDirection, ok := os.LookupEnv("MIGRATION_DIRECTION"); ok && migrateDirection != "" {
		config.MigrateDirection = migrateDirection
		logger.Zap.Info("Loaded DB migration direction from env", "migr_dir", migrateDirection)
	}

	if dbReinitialize, ok := os.LookupEnv("DB_REINIT"); ok {
		config.DBReinitialize = (dbReinitialize == "true")
		logger.Zap.Info("Loaded DB reinitialize flag from env", "db_reinit", dbReinitialize)
	}

	if authorizationKey, ok := os.LookupEnv("AUTH_SECRET_KEY"); ok && authorizationKey != "" {
		config.AuthorizationKey = authorizationKey
		logger.Zap.Info("Loaded authorization key from env", authorizationKey)
	}

	// Тут грязное легаси:

	// if shortIDServPort, ok := os.LookupEnv("BASE_URL"); ok && shortIDServPort != "" {
	// 	config.ShortIDServerPort = shortIDServPort
	// }

	// if saveFilePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && saveFilePath != "" {
	// 	config.SaveFilePath = saveFilePath
	// 	logger.Zap.Infof("EnvFilePath = %s", saveFilePath)
	// }

	// if logVerbose, ok := os.LookupEnv("LOGGER_VERBOSE"); ok {
	// 	config.LogVerbose = (logVerbose == "true")
	// 	logger.Zap.Infof("EnvVerbose = %s", logVerbose)
	// }

	// // Проверим режим работы с данными и сформируем соотвествующий индикатор,
	// // проверять будем по порядку:
	// if config.DatabaseURI != "" {
	// 	config.MemMode = "postgres"
	// } else if config.SaveFilePath != "" {
	// 	config.MemMode = "savefile"
	// } else {
	// 	config.MemMode = "memstore"
	// }

	// var useFile bool
	// var useDSN bool

	logger.Zap.Infof("config.DatabaseURI: %s. DatabaseURI: %s.",
		config.DatabaseURI, databaseURI)

	return config
}

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
