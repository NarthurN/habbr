package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config представляет полную конфигурацию приложения Habbr.
//
// Структура объединяет все разделы конфигурации и загружается автоматически
// из переменных окружения с помощью библиотеки envconfig. Поддерживает
// валидацию всех параметров и предоставляет удобные методы для доступа.
//
// Конфигурация разделена на логические секции:
// - Server: настройки HTTP сервера и GraphQL API
// - Database: параметры подключения к базе данных
// - Logger: настройки системы логирования
//
// Пример использования:
//   cfg, err := config.Load()
//   if err != nil {
//       log.Fatal("Failed to load config:", err)
//   }
//
//   fmt.Printf("Server will run on %s\n", cfg.GetServerAddress())
//   if cfg.IsPostgresDatabase() {
//       fmt.Println("Using PostgreSQL database")
//   }
type Config struct {
	// Server содержит настройки HTTP сервера и GraphQL API
	Server ServerConfig `envconfig:"SERVER"`

	// Database содержит параметры подключения к базе данных
	Database DatabaseConfig `envconfig:"DATABASE"`

	// Logger содержит настройки системы логирования
	Logger LoggerConfig `envconfig:"LOGGER"`
}

// ServerConfig содержит настройки HTTP сервера и GraphQL API.
//
// Определяет параметры сетевого взаимодействия, таймауты, и возможности API.
// Все настройки имеют разумные значения по умолчанию для разработки.
//
// Переменные окружения имеют префикс SERVER_, например:
//   SERVER_HOST=0.0.0.0
//   SERVER_PORT=8080
//   SERVER_READ_TIMEOUT=30s
//   SERVER_ENABLE_PLAYGROUND=true
//
// Пример использования:
//   if cfg.Server.EnablePlayground {
//       fmt.Println("GraphQL Playground доступен по адресу:", cfg.GetServerAddress())
//   }
type ServerConfig struct {
	// Host - IP адрес для привязки сервера
	// Значение по умолчанию: "localhost"
	// Для Docker используйте "0.0.0.0"
	Host string `envconfig:"HOST" default:"localhost"`

	// Port - порт для HTTP сервера
	// Значение по умолчанию: 8080
	// Диапазон: 1-65535
	Port int `envconfig:"PORT" default:"8080"`

	// ReadTimeout - максимальное время чтения запроса
	// Значение по умолчанию: 30s
	// Защищает от медленных клиентов
	ReadTimeout time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`

	// WriteTimeout - максимальное время записи ответа
	// Значение по умолчанию: 30s
	// Защищает от медленных клиентов
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`

	// IdleTimeout - время жизни idle соединений
	// Значение по умолчанию: 120s
	// Помогает освобождать ресурсы
	IdleTimeout time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`

	// ShutdownTimeout - время ожидания корректного завершения
	// Значение по умолчанию: 30s
	// Время для завершения активных запросов при остановке
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`

	// EnablePlayground - включить GraphQL Playground
	// Значение по умолчанию: true
	// В продакшене рекомендуется отключать (false)
	EnablePlayground bool `envconfig:"ENABLE_PLAYGROUND" default:"true"`

	// EnableIntrospection - включить GraphQL introspection
	// Значение по умолчанию: true
	// В продакшене рекомендуется отключать (false)
	EnableIntrospection bool `envconfig:"ENABLE_INTROSPECTION" default:"true"`
}

// DatabaseConfig содержит настройки подключения к базе данных.
//
// Поддерживает два типа хранилища:
// - "postgres": PostgreSQL база данных для продакшена
// - "memory": In-memory хранилище для разработки и тестов
//
// Переменные окружения имеют префикс DATABASE_, например:
//   DATABASE_TYPE=postgres
//   DATABASE_HOST=localhost
//   DATABASE_PORT=5432
//   DATABASE_NAME=habbr
//   DATABASE_USER=habbr_user
//   DATABASE_PASSWORD=secret
//
// Пример использования:
//   if cfg.Database.Type == "postgres" {
//       connectionString := cfg.GetPostgresConnectionString()
//       db, err := sql.Open("postgres", connectionString)
//   }
type DatabaseConfig struct {
	// Type - тип базы данных
	// Значения: "postgres", "memory"
	// Значение по умолчанию: "memory"
	Type string `envconfig:"TYPE" default:"memory"`

	// Host - хост PostgreSQL сервера
	// Значение по умолчанию: "localhost"
	// Игнорируется для типа "memory"
	Host string `envconfig:"HOST" default:"localhost"`

	// Port - порт PostgreSQL сервера
	// Значение по умолчанию: 5432
	// Игнорируется для типа "memory"
	Port int `envconfig:"PORT" default:"5432"`

	// Name - имя базы данных PostgreSQL
	// Значение по умолчанию: "habbr"
	// Игнорируется для типа "memory"
	Name string `envconfig:"NAME" default:"habbr"`

	// User - имя пользователя PostgreSQL
	// Значение по умолчанию: "habbr"
	// Игнорируется для типа "memory"
	User string `envconfig:"USER" default:"habbr"`

	// Password - пароль пользователя PostgreSQL
	// Значение по умолчанию: "password"
	// Игнорируется для типа "memory"
	Password string `envconfig:"PASSWORD" default:"password"`

	// SSLMode - режим SSL для PostgreSQL
	// Значения: "disable", "require", "verify-ca", "verify-full"
	// Значение по умолчанию: "disable"
	SSLMode string `envconfig:"SSL_MODE" default:"disable"`

	// MaxConnections - максимальное количество соединений в пуле
	// Значение по умолчанию: 25
	// Рекомендуется настраивать под нагрузку
	MaxConnections int `envconfig:"MAX_CONNECTIONS" default:"25"`

	// MaxIdleTime - максимальное время жизни idle соединения
	// Значение по умолчанию: 30m
	// Помогает освобождать ресурсы БД
	MaxIdleTime time.Duration `envconfig:"MAX_IDLE_TIME" default:"30m"`

	// MaxLifetime - максимальное время жизни соединения
	// Значение по умолчанию: 2h
	// Предотвращает накопление проблемных соединений
	MaxLifetime time.Duration `envconfig:"MAX_LIFETIME" default:"2h"`
}

// LoggerConfig содержит настройки системы логирования.
//
// Управляет уровнем детализации логов, форматом вывода и дополнительной
// информацией. Поддерживает структурированное логирование через zap.
//
// Переменные окружения имеют префикс LOGGER_, например:
//   LOGGER_LEVEL=info
//   LOGGER_FORMAT=json
//   LOGGER_ENABLE_CALLER=true
//
// Пример использования:
//   if cfg.Logger.Level == "debug" {
//       fmt.Println("Включен отладочный режим логирования")
//   }
type LoggerConfig struct {
	// Level - уровень логирования
	// Значения: "debug", "info", "warn", "error"
	// Значение по умолчанию: "info"
	// debug: самый подробный, error: только ошибки
	Level string `envconfig:"LEVEL" default:"info"`

	// Format - формат вывода логов
	// Значения: "json", "console"
	// Значение по умолчанию: "json"
	// json: структурированный для продакшена, console: читаемый для разработки
	Format string `envconfig:"FORMAT" default:"json"`

	// EnableCaller - включить информацию о вызывающем коде
	// Значение по умолчанию: true
	// Добавляет имя файла и номер строки в логи
	EnableCaller bool `envconfig:"ENABLE_CALLER" default:"true"`
}

// Load загружает конфигурацию из переменных окружения с валидацией.
//
// Функция использует библиотеку envconfig для автоматического сканирования
// переменных окружения и заполнения структуры конфигурации. После загрузки
// выполняется полная валидация всех параметров.
//
// Возвращает:
//   - *Config: полностью загруженная и валидированная конфигурация
//   - error: ошибка загрузки или валидации
//
// Возможные ошибки:
//   - Ошибки парсинга переменных окружения
//   - Ошибки валидации значений (неверный порт, уровень логирования и т.д.)
//   - Отсутствие обязательных параметров для PostgreSQL
//
// Примеры переменных окружения:
//   export SERVER_PORT=8080
//   export DATABASE_TYPE=postgres
//   export DATABASE_HOST=db.example.com
//   export DATABASE_USER=myuser
//   export DATABASE_PASSWORD=mypassword
//   export LOGGER_LEVEL=debug
//
// Пример использования:
//   cfg, err := config.Load()
//   if err != nil {
//       log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
//   }
//
//   fmt.Printf("Сервер запустится на %s\n", cfg.GetServerAddress())
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate выполняет комплексную валидацию всех параметров конфигурации.
//
// Метод проверяет корректность значений во всех секциях конфигурации:
// - Порт сервера в допустимом диапазоне (1-65535)
// - Корректный тип базы данных ("postgres" или "memory")
// - Наличие обязательных параметров для PostgreSQL
// - Валидный уровень логирования
// - Валидный формат логирования
//
// Возвращает:
//   - nil: если вся конфигурация корректна
//   - error: описание первой найденной ошибки валидации
//
// Типичные ошибки валидации:
//   - "invalid server port: 70000" - порт вне диапазона
//   - "invalid database type: mysql" - неподдерживаемый тип БД
//   - "database host is required for postgres" - отсутствует хост для PostgreSQL
//   - "invalid logger level: trace" - неподдерживаемый уровень логирования
//
// Пример использования:
//   cfg := &Config{...}
//   if err := cfg.Validate(); err != nil {
//       return fmt.Errorf("конфигурация невалидна: %w", err)
//   }
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Type != "postgres" && c.Database.Type != "memory" {
		return fmt.Errorf("invalid database type: %s (must be 'postgres' or 'memory')", c.Database.Type)
	}

	if c.Database.Type == "postgres" {
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for postgres")
		}
		if c.Database.Name == "" {
			return fmt.Errorf("database name is required for postgres")
		}
		if c.Database.User == "" {
			return fmt.Errorf("database user is required for postgres")
		}
	}

	if c.Logger.Level != "debug" && c.Logger.Level != "info" &&
		c.Logger.Level != "warn" && c.Logger.Level != "error" {
		return fmt.Errorf("invalid logger level: %s", c.Logger.Level)
	}

	if c.Logger.Format != "json" && c.Logger.Format != "console" {
		return fmt.Errorf("invalid logger format: %s", c.Logger.Format)
	}

	return nil
}

// GetServerAddress возвращает полный сетевой адрес сервера для привязки.
//
// Комбинирует хост и порт в формате "host:port", который может быть
// использован для запуска HTTP сервера.
//
// Возвращает:
//   - string: адрес в формате "host:port"
//
// Примеры возвращаемых значений:
//   - "localhost:8080" - для разработки
//   - "0.0.0.0:8080" - для Docker
//   - "192.168.1.100:3000" - для кастомной конфигурации
//
// Пример использования:
//   address := cfg.GetServerAddress()
//   server := &http.Server{
//       Addr:    address,
//       Handler: handler,
//   }
//   fmt.Printf("Сервер запускается на %s\n", address)
//   if err := server.ListenAndServe(); err != nil {
//       log.Fatal(err)
//   }
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetPostgresConnectionString формирует строку подключения к PostgreSQL.
//
// Создает DSN (Data Source Name) строку в формате, совместимом с
// драйвером pq и pgx. Включает все необходимые параметры подключения.
//
// Возвращает:
//   - string: строка подключения PostgreSQL
//
// Формат строки: "host=HOST port=PORT dbname=DB user=USER password=PASS sslmode=MODE"
//
// Примеры возвращаемых значений:
//   - "host=localhost port=5432 dbname=habbr user=habbr password=password sslmode=disable"
//   - "host=db.example.com port=5433 dbname=prod_habbr user=app_user password=secret123 sslmode=require"
//
// Примечание: Метод НЕ проверяет, что тип базы данных равен "postgres".
// Вызывающий код должен убедиться в этом заранее.
//
// Пример использования:
//   if cfg.IsPostgresDatabase() {
//       dsn := cfg.GetPostgresConnectionString()
//       pool, err := pgxpool.Connect(ctx, dsn)
//       if err != nil {
//           return fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
//       }
//       defer pool.Close()
//   }
func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.User,
		c.Database.Password,
		c.Database.SSLMode,
	)
}

// IsPostgresDatabase проверяет, настроено ли приложение для использования PostgreSQL.
//
// Удобный метод для проверки типа базы данных без прямого обращения к полю.
// Используется для условного выполнения PostgreSQL-специфичного кода.
//
// Возвращает:
//   - true: если Database.Type равен "postgres"
//   - false: если используется другой тип базы данных
//
// Пример использования:
//   if cfg.IsPostgresDatabase() {
//       // Инициализировать PostgreSQL подключение
//       repo, err := postgres.NewRepository(cfg.GetPostgresConnectionString())
//       if err != nil {
//           return err
//       }
//   } else {
//       // Использовать in-memory репозиторий
//       repo := memory.NewRepository()
//   }
func (c *Config) IsPostgresDatabase() bool {
	return c.Database.Type == "postgres"
}

// IsMemoryDatabase проверяет, настроено ли приложение для использования in-memory хранилища.
//
// Удобный метод для проверки типа базы данных без прямого обращения к полю.
// In-memory хранилище используется для разработки, тестирования и демонстрации.
//
// Возвращает:
//   - true: если Database.Type равен "memory"
//   - false: если используется другой тип базы данных
//
// Пример использования:
//   if cfg.IsMemoryDatabase() {
//       fmt.Println("ВНИМАНИЕ: Используется временное хранилище в памяти")
//       fmt.Println("Все данные будут потеряны при перезапуске приложения")
//
//       // Инициализировать in-memory репозиторий
//       repo := memory.NewRepository()
//   }
func (c *Config) IsMemoryDatabase() bool {
	return c.Database.Type == "memory"
}
