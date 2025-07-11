package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"

	"github.com/NarthurN/habbr/internal/api/graphql/generated"
	"github.com/NarthurN/habbr/internal/api/graphql/resolver"
	"github.com/NarthurN/habbr/internal/config"
	"github.com/NarthurN/habbr/internal/repository"
	"github.com/NarthurN/habbr/internal/repository/memory"
	"github.com/NarthurN/habbr/internal/service"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Настройка логгера
	logger, err := setupLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Habbr GraphQL API server",
		zap.String("version", "1.0.0"),
		zap.String("database_type", cfg.Database.Type),
		zap.String("server_address", cfg.GetServerAddress()),
	)

	// Инициализация репозиториев
	repoManager, err := setupRepositories(cfg)
	if err != nil {
		logger.Fatal("Failed to setup repositories", zap.Error(err))
	}
	defer func() {
		if err := repoManager.Close(context.Background()); err != nil {
			logger.Error("Failed to close repositories", zap.Error(err))
		}
	}()

	// Инициализация сервисов
	serviceManager := service.NewManager(repoManager.GetRepositories(), logger)
	defer serviceManager.Close()

	// Настройка GraphQL сервера
	srv := setupGraphQLServer(cfg, serviceManager.GetServices(), logger)

	// Настройка HTTP сервера
	httpServer := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      setupHTTPHandlers(cfg, srv),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", cfg.GetServerAddress()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	waitForShutdown(logger, httpServer, cfg.Server.ShutdownTimeout)
}

// setupLogger настраивает логгер на основе конфигурации
func setupLogger(cfg config.LoggerConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch cfg.Format {
	case "console":
		zapConfig = zap.NewDevelopmentConfig()
	default: // json
		zapConfig = zap.NewProductionConfig()
	}

	// Установка уровня логирования
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	zapConfig.DisableCaller = !cfg.EnableCaller

	return zapConfig.Build()
}

// setupRepositories инициализирует репозитории на основе конфигурации
func setupRepositories(cfg *config.Config) (interface {
	GetRepositories() *repository.Repositories
	Close(context.Context) error
}, error) {
	switch cfg.Database.Type {
	case "memory":
		return memory.NewManager(), nil
	case "postgres":
		// TODO: Реализовать PostgreSQL менеджер
		return nil, fmt.Errorf("PostgreSQL repository not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}

// setupGraphQLServer настраивает GraphQL сервер
func setupGraphQLServer(cfg *config.Config, services *service.Services, logger *zap.Logger) *handler.Server {
	// Создаем резолвер с внедренными зависимостями
	resolverImpl := resolver.NewResolver(services, logger.Named("graphql"))

	// Создаем исполняемую схему
	executableSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolverImpl,
	})

	// Создаем сервер
	srv := handler.New(executableSchema)

	// Добавляем транспорты
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Настройка кэширования
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Добавляем расширения
	if cfg.Server.EnableIntrospection {
		srv.Use(extension.Introspection{})
	}

	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	logger.Info("GraphQL server configured successfully",
		zap.Bool("introspection", cfg.Server.EnableIntrospection),
		zap.Bool("playground", cfg.Server.EnablePlayground),
	)

	return srv
}

// setupHTTPHandlers настраивает HTTP обработчики
func setupHTTPHandlers(cfg *config.Config, graphqlServer *handler.Server) http.Handler {
	mux := http.NewServeMux()

	// GraphQL endpoint
	mux.Handle("/query", graphqlServer)

	// GraphQL Playground (только в режиме разработки)
	if cfg.Server.EnablePlayground {
		mux.Handle("/", playground.Handler("Habbr GraphQL Playground", "/query"))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"service":"habbr-graphql-api","status":"running","endpoints":["/query","/health"]}`)
		})
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"habbr-graphql-api","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Metrics endpoint (базовый)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"service":"habbr-graphql-api","uptime":"unknown"}`)
	})

	return mux
}

// waitForShutdown ожидает сигнал завершения и выполняет graceful shutdown
func waitForShutdown(logger *zap.Logger, server *http.Server, timeout time.Duration) {
	// Канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала
	sig := <-quit
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Остановка HTTP сервера
	logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("HTTP server gracefully stopped")
	}
}
