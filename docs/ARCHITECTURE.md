# Архитектура Habbr GraphQL API

## Обзор архитектуры

Habbr построен с использованием принципов **Clean Architecture** (Чистая Архитектура), которая обеспечивает:

- **Независимость от фреймворков**: Архитектура не зависит от внешних библиотек
- **Тестируемость**: Бизнес-логика легко тестируется в изоляции
- **Независимость от UI**: GraphQL API может быть заменен на REST или gRPC
- **Независимость от БД**: Легко переключаться между PostgreSQL и in-memory
- **Независимость от внешних сервисов**: Бизнес-логика изолирована

## Диаграмма слоев

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   GraphQL       │  │   Resolvers     │  │ Converters  │  │
│  │   Schemas       │  │                 │  │             │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   Post Service  │  │ Comment Service │  │ Subscription│  │
│  │                 │  │                 │  │  Service    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   Interfaces    │  │   PostgreSQL    │  │  In-Memory  │  │
│  │                 │  │  Implementation │  │ Implementation│  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │     Models      │  │   Validation    │  │   Business  │  │
│  │                 │  │     Rules       │  │    Rules    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Детальное описание слоев

### 1. Domain Layer (Доменный слой)

**Расположение**: `internal/model/`

Это центральный слой архитектуры, содержащий бизнес-логику и правила предметной области.

#### Компоненты:

**Доменные модели** (`post.go`, `comment.go`):
```go
type Post struct {
    ID              uuid.UUID
    Title           string
    Content         string
    AuthorID        uuid.UUID
    CommentsEnabled bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

func (p *Post) Validate() error {
    // Бизнес-правила валидации
}
```

**Входные типы** (`input.go`):
```go
type PostInput struct {
    Title           string
    Content         string
    AuthorID        uuid.UUID
    CommentsEnabled bool
}
```

**Типы соединений** (`connection.go`):
```go
type PostConnection struct {
    Edges    []*PostEdge
    PageInfo *PageInfo
}
```

#### Принципы:
- Не зависит от внешних слоев
- Содержит бизнес-правила и валидацию
- Определяет контракты через интерфейсы
- Не содержит технических деталей

### 2. Repository Layer (Слой доступа к данным)

**Расположение**: `internal/repository/`

Обеспечивает абстракцию доступа к данным и реализует паттерн Repository.

#### Структура:

```
repository/
├── interfaces.go          # Интерфейсы репозиториев
├── postgres/             # PostgreSQL реализация
│   ├── post.go
│   ├── comment.go
│   └── connection.go
└── memory/               # In-memory реализация
    ├── post.go
    ├── comment.go
    └── storage.go
```

#### Интерфейсы:

```go
type PostRepository interface {
    GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error)
    ListPosts(ctx context.Context, filter *model.PostFilter, pagination *model.PaginationInput) (*model.PostConnection, error)
    CreatePost(ctx context.Context, input *model.PostInput) (*model.Post, error)
    UpdatePost(ctx context.Context, id uuid.UUID, input *model.PostUpdateInput) (*model.Post, error)
    DeletePost(ctx context.Context, id uuid.UUID) error
}
```

#### PostgreSQL реализация:
- Использует pgx драйвер для производительности
- Оптимизированные SQL запросы с индексами
- Транзакции для консистентности данных
- Connection pooling для масштабируемости

#### In-Memory реализация:
- Для разработки и тестирования
- Полная совместимость с интерфейсом
- Поддержка всех операций включая пагинацию

### 3. Service Layer (Бизнес-логика)

**Расположение**: `internal/service/`

Содержит use cases и координирует взаимодействие между слоями.

#### Компоненты:

**PostService** (`post.go`):
```go
type PostService struct {
    postRepo    repository.PostRepository
    commentRepo repository.CommentRepository
    logger      *zap.Logger
}

func (s *PostService) CreatePost(ctx context.Context, input *model.PostInput) (*model.Post, error) {
    // Валидация
    if err := input.Validate(); err != nil {
        return nil, err
    }

    // Бизнес-логика
    return s.postRepo.CreatePost(ctx, input)
}
```

**CommentService** (`comment.go`):
- Управление иерархическими комментариями
- Проверка глубины вложенности
- Валидация родительских комментариев

**SubscriptionService** (`subscription.go`):
- Pub/Sub для real-time уведомлений
- Управление WebSocket соединениями
- Фильтрация событий по подпискам

#### Принципы:
- Инкапсулирует бизнес-логику
- Координирует операции между репозиториями
- Обрабатывает валидацию и авторизацию
- Не зависит от внешних слоев

### 4. API Layer (Слой представления)

**Расположение**: `internal/api/graphql/`

Предоставляет GraphQL интерфейс для взаимодействия с приложением.

#### Структура:

```
api/graphql/
├── schema/               # GraphQL схемы
│   ├── schema.graphql
│   ├── types.graphql
│   ├── query.graphql
│   ├── mutation.graphql
│   └── subscription.graphql
├── generated/            # Сгенерированный код
├── resolver/             # Резолверы
│   ├── resolver.go
│   ├── query.resolvers.go
│   ├── mutation.resolvers.go
│   └── subscription.resolvers.go
└── converter/           # Конвертеры типов
    ├── post.go
    └── comment.go
```

#### Резолверы:

```go
func (r *queryResolver) Posts(ctx context.Context, first *int, after *string, filter *generated.PostFilter) (*generated.PostConnection, error) {
    // Конвертация GraphQL типов в доменные
    domainFilter, err := converter.PostFilterFromGraphQL(filter)
    if err != nil {
        return nil, err
    }

    // Вызов сервиса
    result, err := r.postService.ListPosts(ctx, domainFilter, pagination)
    if err != nil {
        return nil, err
    }

    // Конвертация обратно в GraphQL типы
    return converter.PostConnectionToGraphQL(result), nil
}
```

#### Конвертеры:
- Преобразование между GraphQL и доменными типами
- Валидация входных данных
- Обработка ошибок

## Паттерны проектирования

### Repository Pattern

Абстрагирует доступ к данным через интерфейсы:

```go
// Интерфейс (contracts)
type PostRepository interface {
    GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error)
    // ...
}

// PostgreSQL реализация
type postgresPostRepository struct {
    db *pgxpool.Pool
}

// In-memory реализация
type memoryPostRepository struct {
    storage map[uuid.UUID]*model.Post
}
```

### Dependency Injection

Внедрение зависимостей через конструкторы:

```go
type PostService struct {
    postRepo    repository.PostRepository
    commentRepo repository.CommentRepository
    logger      *zap.Logger
}

func NewPostService(postRepo repository.PostRepository, commentRepo repository.CommentRepository, logger *zap.Logger) *PostService {
    return &PostService{
        postRepo:    postRepo,
        commentRepo: commentRepo,
        logger:      logger,
    }
}
```

### Publisher-Subscriber

Для real-time уведомлений:

```go
type SubscriptionService struct {
    subscribers map[string]map[string]chan *model.CommentEvent
    mu          sync.RWMutex
}

func (s *SubscriptionService) Subscribe(postID string, clientID string) <-chan *model.CommentEvent {
    // Подписка на события
}

func (s *SubscriptionService) Publish(event *model.CommentEvent) {
    // Публикация события
}
```

## Конфигурация и зависимости

### Инициализация приложения

```go
func main() {
    // Загрузка конфигурации
    cfg := config.Load()

    // Инициализация логгера
    logger := setupLogger(cfg)

    // Инициализация репозиториев
    postRepo := setupPostRepository(cfg, logger)
    commentRepo := setupCommentRepository(cfg, logger)

    // Инициализация сервисов
    postService := service.NewPostService(postRepo, commentRepo, logger)
    commentService := service.NewCommentService(commentRepo, postRepo, logger)
    subscriptionService := service.NewSubscriptionService(logger)

    // Инициализация резолвера
    resolver := resolver.NewResolver(postService, commentService, subscriptionService, logger)

    // Запуск сервера
    startServer(cfg, resolver, logger)
}
```

## Обработка ошибок

### Принципы обработки ошибок

1. **Явная обработка**: Все ошибки обрабатываются явно
2. **Контекст**: Ошибки оборачиваются с контекстом
3. **Логирование**: Все ошибки логируются с контекстом
4. **Типизация**: Использование типизированных ошибок

### Примеры:

```go
// Оборачивание ошибки с контекстом
func (r *postgresPostRepository) GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error) {
    var post model.Post
    err := r.db.QueryRow(ctx, query, id).Scan(...)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, model.ErrPostNotFound
        }
        return nil, fmt.Errorf("failed to get post %s: %w", id, err)
    }
    return &post, nil
}

// Обработка в сервисе
func (s *PostService) GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error) {
    post, err := s.postRepo.GetPost(ctx, id)
    if err != nil {
        s.logger.Error("failed to get post",
            zap.String("postID", id.String()),
            zap.Error(err))
        return nil, err
    }
    return post, nil
}
```

## Тестирование архитектуры

### Unit тесты

Каждый слой тестируется изолированно с использованием моков:

```go
func TestPostService_CreatePost(t *testing.T) {
    // Arrange
    mockRepo := &mocks.PostRepository{}
    service := service.NewPostService(mockRepo, nil, zap.NewNop())

    input := &model.PostInput{
        Title:   "Test Post",
        Content: "Test Content",
    }

    expectedPost := &model.Post{
        ID:      uuid.New(),
        Title:   input.Title,
        Content: input.Content,
    }

    mockRepo.On("CreatePost", mock.Anything, input).Return(expectedPost, nil)

    // Act
    result, err := service.CreatePost(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedPost, result)
    mockRepo.AssertExpectations(t)
}
```

### Integration тесты

Тестирование с реальной базой данных:

```go
func TestPostRepository_Integration(t *testing.T) {
    // Настройка тестовой БД
    db := setupTestDB(t)
    defer db.Close()

    repo := postgres.NewPostRepository(db)

    // Тест создания и получения поста
    input := &model.PostInput{...}
    created, err := repo.CreatePost(context.Background(), input)
    assert.NoError(t, err)

    retrieved, err := repo.GetPost(context.Background(), created.ID)
    assert.NoError(t, err)
    assert.Equal(t, created, retrieved)
}
```

## Производительность и масштабирование

### Оптимизации базы данных

1. **Индексы**: Оптимизированные индексы для частых запросов
2. **Connection Pooling**: Пул соединений для эффективного использования ресурсов
3. **Prepared Statements**: Кэширование планов выполнения
4. **Пагинация**: Cursor-based пагинация для больших наборов данных

### Кэширование

```go
type CachedPostRepository struct {
    repo  repository.PostRepository
    cache *lru.Cache
}

func (c *CachedPostRepository) GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error) {
    if cached, ok := c.cache.Get(id.String()); ok {
        return cached.(*model.Post), nil
    }

    post, err := c.repo.GetPost(ctx, id)
    if err != nil {
        return nil, err
    }

    c.cache.Add(id.String(), post)
    return post, nil
}
```

### Мониторинг

- **Метрики**: Время выполнения операций, количество запросов
- **Логирование**: Структурированные логи с контекстом
- **Трейсинг**: Распределенное трассирование запросов
- **Health Checks**: Проверка состояния всех компонентов

## Заключение

Архитектура Habbr обеспечивает:

- **Модульность**: Четкое разделение ответственности
- **Тестируемость**: Каждый компонент легко тестировать
- **Масштабируемость**: Возможность горизонтального масштабирования
- **Поддерживаемость**: Легкость внесения изменений
- **Расширяемость**: Простота добавления новой функциональности

Эта архитектура позволяет команде эффективно разрабатывать и поддерживать сложные системы, сохраняя при этом высокое качество кода и производительность.
