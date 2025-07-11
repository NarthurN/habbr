package postgres

import (
	"context"
	"fmt"

	"github.com/NarthurN/habbr/internal/config"
	"github.com/NarthurN/habbr/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Manager реализует RepositoryManager для PostgreSQL
type Manager struct {
	pool   *pgxpool.Pool
	config *config.DatabaseConfig
	logger *zap.Logger
	repos  *repository.Repositories
}

// NewManager создает новый менеджер PostgreSQL репозиториев
func NewManager(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// Формируем DSN для подключения
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	// Настраиваем конфигурацию пула
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Настройки пула соединений
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = 1 // Устанавливаем минимум 1 соединение
	poolConfig.MaxConnLifetime = cfg.MaxLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime

	// Настройки логирования для pgx
	if logger != nil {
		poolConfig.ConnConfig.Tracer = &queryTracer{logger: logger}
	}

	// Создаем пул соединений
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	manager := &Manager{
		pool:   pool,
		config: cfg,
		logger: logger,
	}

	// Инициализируем репозитории
	manager.repos = &repository.Repositories{
		Post:    NewPostRepository(pool, logger),
		Comment: NewCommentRepository(pool, logger),
	}

	logger.Info("PostgreSQL manager initialized successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.Int("max_connections", cfg.MaxConnections),
	)

	return manager, nil
}

// GetRepositories возвращает все репозитории
func (m *Manager) GetRepositories() *repository.Repositories {
	return m.repos
}

// Close закрывает все соединения с базой данных
func (m *Manager) Close(ctx context.Context) error {
	if m.pool != nil {
		m.logger.Info("Closing PostgreSQL connection pool")
		m.pool.Close()
	}
	return nil
}

// HealthCheck проверяет состояние соединения с базой данных
func (m *Manager) HealthCheck(ctx context.Context) error {
	if m.pool == nil {
		return fmt.Errorf("connection pool is not initialized")
	}

	// Проверяем доступность базы данных
	if err := m.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Проверяем статистику пула
	stats := m.pool.Stat()
	m.logger.Debug("Database pool stats",
		zap.Int32("total_connections", stats.TotalConns()),
		zap.Int32("idle_connections", stats.IdleConns()),
		zap.Int32("acquired_connections", stats.AcquiredConns()),
	)

	return nil
}

// Migrate выполняет миграции базы данных
func (m *Manager) Migrate(ctx context.Context) error {
	if m.pool == nil {
		return fmt.Errorf("connection pool is not initialized")
	}

	m.logger.Info("Starting database migration")

	// Получаем соединение из пула
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection for migration: %w", err)
	}
	defer conn.Release()

	// Начинаем транзакцию для миграции
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			m.logger.Error("Failed to rollback migration transaction", zap.Error(err))
		}
	}()

	// Создаем таблицу миграций если её нет
	if err := m.createMigrationsTable(ctx, tx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Выполняем миграции
	migrations := []Migration{
		{
			Version:     1,
			Description: "Initial schema with posts and comments",
			SQL: `
				-- Создание таблицы постов
				CREATE TABLE IF NOT EXISTS posts (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					title VARCHAR(200) NOT NULL,
					content TEXT NOT NULL,
					author_id UUID NOT NULL,
					comments_enabled BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Создание индексов для постов
				CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
				CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_posts_comments_enabled ON posts(comments_enabled);

				-- Создание таблицы комментариев
				CREATE TABLE IF NOT EXISTS comments (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					post_id UUID NOT NULL,
					parent_id UUID NULL,
					content TEXT NOT NULL CHECK (LENGTH(content) <= 2000),
					author_id UUID NOT NULL,
					depth INTEGER NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
					FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
				);

				-- Создание индексов для комментариев
				CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
				CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
				CREATE INDEX IF NOT EXISTS idx_comments_author_id ON comments(author_id);
				CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_comments_depth ON comments(depth);

				-- Составной индекс для эффективного получения комментариев к посту
				CREATE INDEX IF NOT EXISTS idx_comments_post_depth_created ON comments(post_id, depth, created_at);

				-- Триггер для автоматического обновления updated_at в постах
				CREATE OR REPLACE FUNCTION update_updated_at_column()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.updated_at = NOW();
					RETURN NEW;
				END;
				$$ language 'plpgsql';

				CREATE TRIGGER update_posts_updated_at
					BEFORE UPDATE ON posts
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

				CREATE TRIGGER update_comments_updated_at
					BEFORE UPDATE ON comments
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
			`,
		},
	}

	for _, migration := range migrations {
		if err := m.runMigration(ctx, tx, migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	m.logger.Info("Database migration completed successfully")
	return nil
}

// Migration представляет миграцию базы данных
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// createMigrationsTable создает таблицу для отслеживания миграций
func (m *Manager) createMigrationsTable(ctx context.Context, tx pgx.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err := tx.Exec(ctx, query)
	return err
}

// runMigration выполняет конкретную миграцию
func (m *Manager) runMigration(ctx context.Context, tx pgx.Tx, migration Migration) error {
	// Проверяем, была ли уже применена эта миграция
	var count int
	err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration.Version).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		m.logger.Debug("Migration already applied", zap.Int("version", migration.Version))
		return nil
	}

	// Выполняем миграцию
	m.logger.Info("Applying migration",
		zap.Int("version", migration.Version),
		zap.String("description", migration.Description),
	)

	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Записываем информацию о примененной миграции
	_, err = tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, description) VALUES ($1, $2)",
		migration.Version, migration.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	m.logger.Info("Migration applied successfully", zap.Int("version", migration.Version))
	return nil
}

// queryTracer реализует pgx.QueryTracer для логирования запросов
type queryTracer struct {
	logger *zap.Logger
}

func (qt *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (qt *queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		qt.logger.Error("Database query failed",
			zap.Error(data.Err),
		)
	} else if qt.logger.Core().Enabled(zap.DebugLevel) {
		qt.logger.Debug("Database query executed")
	}
}
