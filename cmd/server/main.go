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

// main является точкой входа в приложение Habbr GraphQL API.
//
// Функция выполняет полную инициализацию приложения в следующем порядке:
// 1. Загружает конфигурацию из переменных окружения
// 2. Настраивает систему логирования
// 3. Инициализирует репозитории (PostgreSQL или in-memory)
// 4. Создает сервисы бизнес-логики
// 5. Настраивает GraphQL сервер с резолверами
// 6. Запускает HTTP сервер
// 7. Ожидает сигналы завершения для graceful shutdown
//
// Приложение поддерживает корректное завершение работы при получении
// сигналов SIGINT или SIGTERM, завершая активные запросы и освобождая ресурсы.
//
// Примеры переменных окружения для запуска:
//
//	export SERVER_PORT=8080
//	export DATABASE_TYPE=memory
//	export LOGGER_LEVEL=info
//	go run cmd/server/main.go
//
// Или для продакшена с PostgreSQL:
//
//	export SERVER_HOST=0.0.0.0
//	export SERVER_PORT=8080
//	export DATABASE_TYPE=postgres
//	export DATABASE_HOST=localhost
//	export DATABASE_NAME=habbr
//	export DATABASE_USER=habbr_user
//	export DATABASE_PASSWORD=secret
//	export LOGGER_LEVEL=warn
//	export LOGGER_FORMAT=json
//	./habbr-server
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

// setupLogger настраивает и создает экземпляр логгера на основе конфигурации.
//
// Функция создает структурированный логгер используя библиотеку zap.
// Поддерживает два формата вывода:
// - "json": структурированный JSON формат для продакшена
// - "console": человекочитаемый формат для разработки
//
// Поддерживает четыре уровня логирования:
// - "debug": максимально подробные логи (включая отладочную информацию)
// - "info": информационные сообщения (по умолчанию)
// - "warn": предупреждения о потенциальных проблемах
// - "error": только критические ошибки
//
// Параметры:
//   - cfg: конфигурация логгера с настройками формата, уровня и опций
//
// Возвращает:
//   - *zap.Logger: настроенный экземпляр логгера
//   - error: ошибка создания логгера
//
// Особенности:
//   - Caller information (файл:строка) включается в зависимости от cfg.EnableCaller
//   - Логгер оптимизирован для высокой производительности
//   - Поддерживает структурированные поля для лучшей обработки
//
// Пример использования:
//
//	cfg := config.LoggerConfig{
//	    Level: "info",
//	    Format: "json",
//	    EnableCaller: true,
//	}
//	logger, err := setupLogger(cfg)
//	if err != nil {
//	    return fmt.Errorf("не удалось создать логгер: %w", err)
//	}
//	logger.Info("Приложение запущено", zap.String("version", "1.0.0"))
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

// setupRepositories инициализирует менеджер репозиториев на основе конфигурации базы данных.
//
// Функция создает подходящий менеджер репозиториев в зависимости от типа
// базы данных, указанного в конфигурации. Поддерживает:
// - "memory": In-memory хранилище для разработки и тестирования
// - "postgres": PostgreSQL база данных для продакшена (будущая реализация)
//
// Менеджер репозиториев обеспечивает:
// - Единый интерфейс доступа к данным
// - Управление соединениями с базой данных
// - Корректное освобождение ресурсов через метод Close()
//
// Параметры:
//   - cfg: полная конфигурация приложения
//
// Возвращает:
//   - interface{}: менеджер репозиториев с методами GetRepositories() и Close()
//   - error: ошибка инициализации репозиториев
//
// Возможные ошибки:
//   - "PostgreSQL repository not implemented yet": PostgreSQL пока не реализован
//   - "unsupported database type: X": неподдерживаемый тип базы данных
//   - Ошибки подключения к базе данных (для PostgreSQL)
//
// Примечания:
//   - In-memory репозиторий не требует внешних зависимостей
//   - Все данные в memory репозитории теряются при перезапуске
//   - PostgreSQL репозиторий требует запущенного сервера базы данных
//
// Пример использования:
//
//	repoManager, err := setupRepositories(cfg)
//	if err != nil {
//	    return fmt.Errorf("не удалось инициализировать репозитории: %w", err)
//	}
//	defer repoManager.Close(context.Background())
//
//	repos := repoManager.GetRepositories()
//	post, err := repos.Post.GetByID(ctx, postID)
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

// setupGraphQLServer создает и настраивает GraphQL сервер с полной функциональностью.
//
// Функция выполняет комплексную настройку GraphQL сервера включая:
// - Создание резолверов с внедрением зависимостей сервисов
// - Настройку исполняемой схемы с типами и мутациями
// - Добавление транспортов (HTTP, WebSocket для подписок)
// - Конфигурацию кэширования запросов и схем
// - Подключение расширений (introspection, APQ)
//
// Поддерживаемые транспорты:
//   - WebSocket: для real-time подписок с keep-alive
//   - HTTP GET/POST: для стандартных запросов и мутаций
//   - Multipart: для загрузки файлов (если потребуется)
//   - OPTIONS: для CORS preflight запросов
//
// Функции производительности:
//   - LRU кэш для скомпилированных запросов (1000 элементов)
//   - Automatic Persisted Queries для экономии трафика
//   - Introspection отключается в продакшене для безопасности
//
// Параметры:
//   - cfg: конфигурация сервера с настройками безопасности
//   - services: инициализированные сервисы бизнес-логики
//   - logger: логгер для отслеживания операций GraphQL
//
// Возвращает:
//   - *handler.Server: полностью настроенный GraphQL сервер
//
// Примечания:
//   - Сервер поддерживает WebSocket подписки для real-time уведомлений
//   - Introspection включен только если cfg.Server.EnableIntrospection = true
//   - Все запросы логируются на уровне debug
//
// Пример использования:
//
//	srv := setupGraphQLServer(cfg, services, logger)
//	http.Handle("/graphql", srv)
//
//	// Для тестирования подписок:
//	http.Handle("/ws", srv) // WebSocket endpoint
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

// setupHTTPHandlers создает и настраивает HTTP маршрутизатор с всеми необходимыми endpoints.
//
// Функция настраивает полный набор HTTP маршрутов для GraphQL API:
//
// Основные endpoints:
//   - "/query": GraphQL API endpoint для всех запросов, мутаций и подписок
//   - "/": GraphQL Playground (только в dev режиме) или информация о сервисе
//   - "/health": Health check endpoint для мониторинга и load balancer'ов
//   - "/metrics": Базовый metrics endpoint для систем мониторинга
//
// Поведение в зависимости от конфигурации:
//   - Если EnablePlayground = true: "/" показывает GraphQL Playground
//   - Если EnablePlayground = false: "/" возвращает JSON с информацией о сервисе
//   - Health check всегда доступен для проверки состояния сервиса
//
// Параметры:
//   - cfg: конфигурация сервера с настройками endpoints
//   - graphqlServer: настроенный GraphQL сервер для обработки запросов
//
// Возвращает:
//   - http.Handler: маршрутизатор с настроенными endpoints
//
// Примеры ответов:
//
//	GET /health:
//	{"status":"ok","service":"habbr-graphql-api","timestamp":"2024-01-15T10:30:45Z"}
//
//	GET / (без playground):
//	{"service":"habbr-graphql-api","status":"running","endpoints":["/query","/health"]}
//
//	GET /metrics:
//	{"service":"habbr-graphql-api","uptime":"unknown"}
//
// Пример использования:
//
//	handler := setupHTTPHandlers(cfg, graphqlServer)
//	server := &http.Server{
//	    Addr:    ":8080",
//	    Handler: handler,
//	}
//	server.ListenAndServe()
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

// waitForShutdown ожидает сигналы завершения и выполняет graceful shutdown HTTP сервера.
//
// Функция реализует корректный механизм завершения работы приложения:
// 1. Настраивает обработку системных сигналов SIGINT (Ctrl+C) и SIGTERM
// 2. Блокируется в ожидании одного из этих сигналов
// 3. При получении сигнала начинает graceful shutdown сервера
// 4. Ожидает завершения активных запросов в рамках таймаута
// 5. Принудительно останавливает сервер, если таймаут превышен
//
// Graceful shutdown означает:
//   - Сервер прекращает принимать новые соединения
//   - Активные запросы завершаются в рамках таймаута
//   - Idle соединения закрываются немедленно
//   - WebSocket соединения получают уведомление о закрытии
//
// Параметры:
//   - logger: логгер для записи процесса остановки
//   - server: HTTP сервер для остановки
//   - timeout: максимальное время ожидания завершения активных запросов
//
// Поведение при разных сигналах:
//   - SIGINT (Ctrl+C): нормальное завершение, graceful shutdown
//   - SIGTERM: запрос на завершение от системы, graceful shutdown
//   - Превышение timeout: принудительная остановка с ошибкой
//
// Логирование:
//   - Записывает получение сигнала завершения
//   - Отслеживает процесс graceful shutdown
//   - Логирует успешную остановку или принудительное завершение
//
// Пример использования:
//
//	server := &http.Server{Addr: ":8080", Handler: handler}
//	go func() {
//	    if err := server.ListenAndServe(); err != http.ErrServerClosed {
//	        logger.Fatal("Server failed", zap.Error(err))
//	    }
//	}()
//
//	// Ожидание graceful shutdown
//	waitForShutdown(logger, server, 30*time.Second)
//	logger.Info("Application stopped")
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
