package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config представляет конфигурацию приложения
type Config struct {
	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Logger   LoggerConfig   `envconfig:"LOGGER"`
}

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Host                string        `envconfig:"HOST" default:"localhost"`
	Port                int           `envconfig:"PORT" default:"8080"`
	ReadTimeout         time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout        time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout         time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout     time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	EnablePlayground    bool          `envconfig:"ENABLE_PLAYGROUND" default:"true"`
	EnableIntrospection bool          `envconfig:"ENABLE_INTROSPECTION" default:"true"`
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	Type           string        `envconfig:"TYPE" default:"memory"` // "postgres" или "memory"
	Host           string        `envconfig:"HOST" default:"localhost"`
	Port           int           `envconfig:"PORT" default:"5432"`
	Name           string        `envconfig:"NAME" default:"habbr"`
	User           string        `envconfig:"USER" default:"habbr"`
	Password       string        `envconfig:"PASSWORD" default:"password"`
	SSLMode        string        `envconfig:"SSL_MODE" default:"disable"`
	MaxConnections int           `envconfig:"MAX_CONNECTIONS" default:"25"`
	MaxIdleTime    time.Duration `envconfig:"MAX_IDLE_TIME" default:"30m"`
	MaxLifetime    time.Duration `envconfig:"MAX_LIFETIME" default:"2h"`
}

// LoggerConfig содержит настройки логгирования
type LoggerConfig struct {
	Level        string `envconfig:"LEVEL" default:"info"`
	Format       string `envconfig:"FORMAT" default:"json"` // "json" или "console"
	EnableCaller bool   `envconfig:"ENABLE_CALLER" default:"true"`
}

// Load загружает конфигурацию из переменных окружения
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

// Validate проверяет валидность конфигурации
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

// GetServerAddress возвращает адрес сервера
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetPostgresConnectionString возвращает строку подключения к PostgreSQL
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

// IsPostgresDatabase проверяет, используется ли PostgreSQL
func (c *Config) IsPostgresDatabase() bool {
	return c.Database.Type == "postgres"
}

// IsMemoryDatabase проверяет, используется ли in-memory хранилище
func (c *Config) IsMemoryDatabase() bool {
	return c.Database.Type == "memory"
}
