# Документация API Habbr GraphQL

## Обзор

Этот документ содержит полную документацию всех публичных интерфейсов, структур, методов и функций системы Habbr GraphQL API. Документация структурирована по слоям архитектуры для удобства навигации.

## Структура архитектуры

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                            │
│                    GraphQL Resolvers                        │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                           │
│                  Business Logic                             │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                          │
│                   Data Access                               │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│                 Models & Business Rules                     │
└─────────────────────────────────────────────────────────────┘
```

## Domain Layer (Доменный слой)

### Модель Post

#### Структура Post
```go
type Post struct {
    ID              uuid.UUID // Уникальный идентификатор поста
    Title           string    // Заголовок поста, максимум 200 символов
    Content         string    // Содержимое поста, максимум 50000 символов
    AuthorID        uuid.UUID // Идентификатор автора поста
    CommentsEnabled bool      // Флаг разрешения комментариев
    CreatedAt       time.Time // Время создания поста
    UpdatedAt       time.Time // Время последнего обновления
}
```

**Назначение**: Основная сущность системы, представляющая публикацию пользователя с возможностью комментирования.

#### Входные структуры

##### PostInput
```go
type PostInput struct {
    Title           string    // Заголовок поста, обязательное поле
    Content         string    // Содержимое поста, обязательное поле
    AuthorID        uuid.UUID // Идентификатор автора, обязательное поле
    CommentsEnabled bool      // Разрешение комментариев, по умолчанию false
}
```

**Методы**:
- `Validate() error` - валидация входных данных с проверкой длины и обязательности полей

##### PostUpdateInput
```go
type PostUpdateInput struct {
    Title           *string // Новый заголовок поста, опциональное поле
    Content         *string // Новое содержимое поста, опциональное поле
    CommentsEnabled *bool   // Новое значение разрешения комментариев
}
```

**Методы**:
- `Validate() error` - валидация данных обновления

#### Фильтрация и пагинация

##### PostFilter
```go
type PostFilter struct {
    AuthorID     *uuid.UUID // Фильтр по автору поста
    WithComments *bool      // Фильтр по настройке комментариев
}
```

##### PaginationInput
```go
type PaginationInput struct {
    First  *int    // Количество записей с начала (прямая пагинация)
    After  *string // Cursor для продолжения (прямая пагинация)
    Last   *int    // Количество записей с конца (обратная пагинация)
    Before *string // Cursor для продолжения (обратная пагинация)
}
```

#### Connection типы

##### PostConnection
```go
type PostConnection struct {
    Edges    []*PostEdge // Массив ребер с постами и cursors
    PageInfo *PageInfo   // Информация о пагинации
}
```

##### PostEdge
```go
type PostEdge struct {
    Node   *Post  // Сам объект поста
    Cursor string // Уникальный идентификатор позиции
}
```

##### PageInfo
```go
type PageInfo struct {
    HasNextPage     bool    // Есть ли следующая страница
    HasPreviousPage bool    // Есть ли предыдущая страница
    StartCursor     *string // Cursor первого элемента
    EndCursor       *string // Cursor последнего элемента
}
```

#### Функции

##### NewPost
```go
func NewPost(input PostInput) *Post
```
**Назначение**: Создает новый пост с автоматической генерацией ID и временных меток.

**Параметры**:
- `input` - валидированные входные данные

**Возвращает**: Готовый для сохранения пост

**Действия**:
- Генерирует новый UUID
- Обрезает пробелы в текстовых полях
- Устанавливает текущее время для CreatedAt и UpdatedAt

#### Методы Post

##### Update
```go
func (p *Post) Update(input PostUpdateInput)
```
**Назначение**: Обновляет существующий пост новыми данными.

**Параметры**:
- `input` - данные для обновления (поля равные nil игнорируются)

**Побочные эффекты**:
- Изменяет указанные поля поста
- Обновляет UpdatedAt на текущее время

##### CanAddComments
```go
func (p *Post) CanAddComments() bool
```
**Назначение**: Проверяет разрешение добавления комментариев к посту.

**Возвращает**: true если комментарии разрешены, false если запрещены

---

### Модель Comment

#### Структура Comment
```go
type Comment struct {
    ID        uuid.UUID  // Уникальный идентификатор комментария
    PostID    uuid.UUID  // Идентификатор поста
    ParentID  *uuid.UUID // Идентификатор родительского комментария (nil для корневых)
    Content   string     // Содержимое комментария, максимум 2000 символов
    AuthorID  uuid.UUID  // Идентификатор автора комментария
    Depth     int        // Глубина вложенности (0 для корневых)
    CreatedAt time.Time  // Время создания комментария
    UpdatedAt time.Time  // Время последнего обновления
    Children  []*Comment // Дочерние комментарии (заполняется при построении дерева)
}
```

**Назначение**: Представляет комментарий в иерархической системе с неограниченной глубиной вложенности.

#### Входные структуры

##### CommentInput
```go
type CommentInput struct {
    PostID   uuid.UUID  // Идентификатор поста, обязательное поле
    ParentID *uuid.UUID // Идентификатор родительского комментария (опциональное)
    Content  string     // Содержимое комментария, обязательное поле
    AuthorID uuid.UUID  // Идентификатор автора, обязательное поле
}
```

**Методы**:
- `Validate() error` - валидация входных данных

##### CommentUpdateInput
```go
type CommentUpdateInput struct {
    Content *string // Новое содержимое комментария, опциональное поле
}
```

**Методы**:
- `Validate() error` - валидация данных обновления

#### Фильтрация

##### CommentFilter
```go
type CommentFilter struct {
    PostID   *uuid.UUID // Фильтр по посту
    ParentID *uuid.UUID // Фильтр по родительскому комментарию
    AuthorID *uuid.UUID // Фильтр по автору
    MaxDepth *int       // Максимальная глубина вложенности
}
```

#### Connection типы

##### CommentConnection
```go
type CommentConnection struct {
    Edges    []*CommentEdge // Массив ребер с комментариями
    PageInfo *PageInfo      // Информация о пагинации
}
```

##### CommentEdge
```go
type CommentEdge struct {
    Node   *Comment // Сам объект комментария
    Cursor string   // Уникальный идентификатор позиции
}
```

#### Подписки

##### CommentSubscriptionPayload
```go
type CommentSubscriptionPayload struct {
    PostID     uuid.UUID // Идентификатор поста
    Comment    *Comment  // Данные комментария
    ActionType string    // Тип события: "CREATED", "UPDATED", "DELETED"
}
```

**Назначение**: Данные события для real-time подписок на комментарии через WebSocket.

#### Константы

##### MaxCommentLength
```go
const MaxCommentLength = 2000
```
**Назначение**: Максимальная длина содержимого комментария в символах.

#### Функции

##### NewComment
```go
func NewComment(input CommentInput, depth int) *Comment
```
**Назначение**: Создает новый комментарий с автоматической генерацией ID и временных меток.

**Параметры**:
- `input` - валидированные входные данные
- `depth` - глубина вложенности в дереве (0 для корневых)

**Возвращает**: Готовый для сохранения комментарий

##### BuildCommentsTree
```go
func BuildCommentsTree(comments []*Comment) []*Comment
```
**Назначение**: Строит иерархическую древовидную структуру из плоского списка комментариев.

**Параметры**:
- `comments` - плоский список всех комментариев

**Возвращает**: Слайс корневых комментариев с построенным деревом

**Алгоритм**:
1. Создает карту для быстрого поиска по ID
2. Инициализирует пустые массивы Children
3. Распределяет комментарии по родителям
4. Возвращает только корневые комментарии

**Сложность**: O(n), где n - количество комментариев

##### FlattenCommentsTree
```go
func FlattenCommentsTree(tree []*Comment) []*Comment
```
**Назначение**: Преобразует древовидную структуру в плоский упорядоченный список.

**Параметры**:
- `tree` - слайс корневых комментариев с построенным деревом

**Возвращает**: Плоский список в иерархическом порядке (depth-first traversal)

**Применение**: Отображение комментариев в линейном виде с сохранением структуры

#### Методы Comment

##### Update
```go
func (c *Comment) Update(input CommentUpdateInput)
```
**Назначение**: Обновляет существующий комментарий новыми данными.

##### IsRootComment
```go
func (c *Comment) IsRootComment() bool
```
**Назначение**: Проверяет, является ли комментарий корневым (привязан к посту).

**Возвращает**: true если ParentID == nil

##### CanBeRepliedTo
```go
func (c *Comment) CanBeRepliedTo() bool
```
**Назначение**: Проверяет возможность создания ответа на комментарий.

**Возвращает**: В текущей реализации всегда true

##### AddChild
```go
func (c *Comment) AddChild(child *Comment)
```
**Назначение**: Добавляет дочерний комментарий в коллекцию Children.

**Параметры**:
- `child` - дочерний комментарий для добавления

##### GetDepth
```go
func (c *Comment) GetDepth() int
```
**Назначение**: Возвращает глубину вложенности комментария в дереве.

**Возвращает**: Целое число начиная с 0 для корневых комментариев

---

## Service Layer (Сервисный слой)

### PostService

```go
type PostService interface {
    CreatePost(ctx context.Context, input model.PostInput) (*model.Post, error)
    GetPost(ctx context.Context, id uuid.UUID) (*model.Post, error)
    ListPosts(ctx context.Context, filter model.PostFilter, pagination model.PaginationInput) (*model.PostConnection, error)
    UpdatePost(ctx context.Context, id uuid.UUID, input model.PostUpdateInput, authorID uuid.UUID) (*model.Post, error)
    DeletePost(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error
    ToggleComments(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*model.Post, error)
}
```

**Назначение**: Интерфейс сервиса для работы с постами, инкапсулирующий всю бизнес-логику управления постами.

#### Методы

##### CreatePost
**Назначение**: Создает новый пост в системе с полной валидацией.

**Действия**:
- Валидирует входные данные
- Создает пост с уникальным ID
- Сохраняет в репозитории

**Возможные ошибки**:
- `model.ValidationError` - некорректные входные данные
- `model.InternalError` - проблемы с базой данных

##### GetPost
**Назначение**: Получает пост по уникальному идентификатору.

**Возможные ошибки**:
- `model.NotFoundError` - пост не существует
- `model.InternalError` - проблемы с базой данных

##### ListPosts
**Назначение**: Возвращает список постов с фильтрацией и cursor-based пагинацией.

**Особенности**:
- Стабильные результаты при изменении данных
- Поддержка фильтрации по автору и настройкам комментариев

##### UpdatePost
**Назначение**: Обновляет существующий пост с проверкой прав доступа.

**Проверки**:
- Только автор может изменять пост
- Валидация новых данных
- Обновление временной метки

**Возможные ошибки**:
- `model.NotFoundError` - пост не найден
- `model.ForbiddenError` - недостаточно прав
- `model.ValidationError` - некорректные данные

##### DeletePost
**Назначение**: Удаляет пост и все связанные комментарии.

**Особенности**:
- Каскадное удаление комментариев
- Выполнение в транзакции
- Необратимая операция

##### ToggleComments
**Назначение**: Переключает возможность комментирования поста.

**Поведение**:
- Существующие комментарии сохраняются
- Только автор может изменять настройку

---

### CommentService

```go
type CommentService interface {
    CreateComment(ctx context.Context, input model.CommentInput) (*model.Comment, error)
    GetComment(ctx context.Context, id uuid.UUID) (*model.Comment, error)
    ListComments(ctx context.Context, filter model.CommentFilter, pagination model.PaginationInput) (*model.CommentConnection, error)
    UpdateComment(ctx context.Context, id uuid.UUID, input model.CommentUpdateInput, authorID uuid.UUID) (*model.Comment, error)
    DeleteComment(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error
    GetCommentsTree(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error)
    GetCommentStats(ctx context.Context, postID uuid.UUID) (int, error)
}
```

**Назначение**: Интерфейс сервиса для работы с иерархическими комментариями.

#### Методы

##### CreateComment
**Назначение**: Создает новый комментарий или ответ на существующий.

**Проверки и действия**:
- Проверяет существование поста и разрешение комментирования
- Если указан ParentID, проверяет родительский комментарий
- Вычисляет правильную глубину вложенности
- Отправляет real-time уведомление подписчикам

**Возможные ошибки**:
- `model.ValidationError` - некорректные входные данные
- `model.NotFoundError` - пост или родительский комментарий не найден
- `model.ForbiddenError` - комментарии отключены

##### GetComment
**Назначение**: Получает комментарий по идентификатору без дочерних элементов.

##### ListComments
**Назначение**: Возвращает список комментариев с мощной фильтрацией.

**Возможности фильтрации**:
- Все комментарии к определенному посту
- Только корневые комментарии (ParentID = nil)
- Дочерние комментарии определенного родителя
- Комментарии определенного автора
- Ограничение по глубине вложенности

##### UpdateComment
**Назначение**: Обновляет содержимое комментария с проверкой прав.

**Действия**:
- Проверяет права доступа (только автор)
- Валидирует новые данные
- Отправляет уведомление подписчикам

##### DeleteComment
**Назначение**: Удаляет комментарий и все его дочерние комментарии.

**Особенности**:
- Каскадное удаление всех ответов
- Выполнение в транзакции
- Уведомления подписчикам

##### GetCommentsTree
**Назначение**: Возвращает полное дерево комментариев для поста.

**Результат**: Список корневых комментариев с заполненными полями Children

##### GetCommentStats
**Назначение**: Подсчитывает общее количество комментариев к посту.

**Применение**: Отображение счетчиков комментариев

---

### SubscriptionService

```go
type SubscriptionService interface {
    Subscribe(ctx context.Context, postID uuid.UUID) (<-chan *model.CommentSubscriptionPayload, error)
    Publish(postID uuid.UUID, payload *model.CommentSubscriptionPayload)
    GetSubscriberCount(postID uuid.UUID) int
    Shutdown()
}
```

**Назначение**: Интерфейс сервиса для управления real-time подписками на события комментариев.

#### Методы

##### Subscribe
**Назначение**: Создает новую подписку на события комментариев для поста.

**Параметры**:
- `ctx` - контекст подписки, отмена закрывает канал
- `postID` - идентификатор поста для подписки

**Возвращает**: Канал для получения событий

**Жизненный цикл**:
- Подписка активна до отмены контекста
- Канал закрывается автоматически
- Ресурсы освобождаются автоматически

##### Publish
**Назначение**: Отправляет событие всем подписчикам поста.

**Поведение**:
- Отправка происходит асинхронно
- Заблокированные подписчики пропускаются
- Неактивные подписки удаляются
- Метод не блокируется

##### GetSubscriberCount
**Назначение**: Возвращает количество активных подписчиков для поста.

**Применение**: Мониторинг и отладка

##### Shutdown
**Назначение**: Корректно завершает работу сервиса подписок.

**Действия**:
- Закрывает все активные каналы
- Освобождает внутренние ресурсы
- Завершает фоновые горутины
- Блокируется до полного завершения

---

### Services

```go
type Services struct {
    Post         PostService         // Сервис для работы с постами
    Comment      CommentService      // Сервис для работы с комментариями
    Subscription SubscriptionService // Сервис для управления подписками
}
```

**Назначение**: Объединяет все сервисы приложения для передачи в слои представления.

---

## Configuration (Конфигурация)

### Config

```go
type Config struct {
    Server   ServerConfig   // Настройки HTTP сервера и GraphQL API
    Database DatabaseConfig // Параметры подключения к базе данных
    Logger   LoggerConfig   // Настройки системы логирования
}
```

**Назначение**: Полная конфигурация приложения, загружаемая из переменных окружения.

#### Методы

##### Load
```go
func Load() (*Config, error)
```
**Назначение**: Загружает конфигурацию из переменных окружения с валидацией.

**Возможные ошибки**:
- Ошибки парсинга переменных окружения
- Ошибки валидации значений
- Отсутствие обязательных параметров для PostgreSQL

##### Validate
```go
func (c *Config) Validate() error
```
**Назначение**: Выполняет комплексную валидацию всех параметров конфигурации.

**Проверки**:
- Порт сервера в диапазоне 1-65535
- Корректный тип базы данных
- Обязательные параметры для PostgreSQL
- Валидные уровни и форматы логирования

##### GetServerAddress
```go
func (c *Config) GetServerAddress() string
```
**Назначение**: Возвращает сетевой адрес сервера в формате "host:port".

##### GetPostgresConnectionString
```go
func (c *Config) GetPostgresConnectionString() string
```
**Назначение**: Формирует строку подключения к PostgreSQL.

**Формат**: "host=HOST port=PORT dbname=DB user=USER password=PASS sslmode=MODE"

##### IsPostgresDatabase
```go
func (c *Config) IsPostgresDatabase() bool
```
**Назначение**: Проверяет, настроено ли приложение для использования PostgreSQL.

##### IsMemoryDatabase
```go
func (c *Config) IsMemoryDatabase() bool
```
**Назначение**: Проверяет, настроено ли приложение для использования in-memory хранилища.

---

### ServerConfig

```go
type ServerConfig struct {
    Host                string        // IP адрес для привязки сервера
    Port                int           // Порт для HTTP сервера
    ReadTimeout         time.Duration // Максимальное время чтения запроса
    WriteTimeout        time.Duration // Максимальное время записи ответа
    IdleTimeout         time.Duration // Время жизни idle соединений
    ShutdownTimeout     time.Duration // Время ожидания корректного завершения
    EnablePlayground    bool          // Включить GraphQL Playground
    EnableIntrospection bool          // Включить GraphQL introspection
}
```

**Назначение**: Настройки HTTP сервера и GraphQL API.

**Переменные окружения**: PREFIX = "SERVER_"

---

### DatabaseConfig

```go
type DatabaseConfig struct {
    Type           string        // Тип базы данных: "postgres", "memory"
    Host           string        // Хост PostgreSQL сервера
    Port           int           // Порт PostgreSQL сервера
    Name           string        // Имя базы данных PostgreSQL
    User           string        // Имя пользователя PostgreSQL
    Password       string        // Пароль пользователя PostgreSQL
    SSLMode        string        // Режим SSL для PostgreSQL
    MaxConnections int           // Максимальное количество соединений в пуле
    MaxIdleTime    time.Duration // Максимальное время жизни idle соединения
    MaxLifetime    time.Duration // Максимальное время жизни соединения
}
```

**Назначение**: Настройки подключения к базе данных.

**Переменные окружения**: PREFIX = "DATABASE_"

**Поддерживаемые типы**:
- "postgres": PostgreSQL база данных для продакшена
- "memory": In-memory хранилище для разработки и тестов

---

### LoggerConfig

```go
type LoggerConfig struct {
    Level        string // Уровень логирования: "debug", "info", "warn", "error"
    Format       string // Формат вывода: "json", "console"
    EnableCaller bool   // Включить информацию о вызывающем коде
}
```

**Назначение**: Настройки системы логирования.

**Переменные окружения**: PREFIX = "LOGGER_"

---

## Application Entry Point (Точка входа)

### main

```go
func main()
```
**Назначение**: Точка входа в приложение Habbr GraphQL API.

**Порядок инициализации**:
1. Загружает конфигурацию из переменных окружения
2. Настраивает систему логирования
3. Инициализирует репозитории (PostgreSQL или in-memory)
4. Создает сервисы бизнес-логики
5. Настраивает GraphQL сервер с резолверами
6. Запускает HTTP сервер
7. Ожидает сигналы завершения для graceful shutdown

### setupLogger

```go
func setupLogger(cfg config.LoggerConfig) (*zap.Logger, error)
```
**Назначение**: Настраивает структурированный логгер на основе конфигурации.

**Поддерживаемые форматы**:
- "json": структурированный JSON для продакшена
- "console": человекочитаемый для разработки

**Уровни логирования**: debug, info, warn, error

### setupRepositories

```go
func setupRepositories(cfg *config.Config) (interface{...}, error)
```
**Назначение**: Инициализирует менеджер репозиториев на основе типа базы данных.

**Поддерживаемые типы**:
- "memory": In-memory хранилище
- "postgres": PostgreSQL база данных (будущая реализация)

### setupGraphQLServer

```go
func setupGraphQLServer(cfg *config.Config, services *service.Services, logger *zap.Logger) *handler.Server
```
**Назначение**: Создает GraphQL сервер с полной функциональностью.

**Функции**:
- Резолверы с внедрением зависимостей
- Транспорты (HTTP, WebSocket для подписок)
- Кэширование запросов (LRU, 1000 элементов)
- Automatic Persisted Queries
- Introspection (отключается в продакшене)

### setupHTTPHandlers

```go
func setupHTTPHandlers(cfg *config.Config, graphqlServer *handler.Server) http.Handler
```
**Назначение**: Настраивает HTTP маршрутизатор с endpoints.

**Endpoints**:
- `/query` - GraphQL API endpoint
- `/` - GraphQL Playground или информация о сервисе
- `/health` - Health check для мониторинга
- `/metrics` - Базовый metrics endpoint

### waitForShutdown

```go
func waitForShutdown(logger *zap.Logger, server *http.Server, timeout time.Duration)
```
**Назначение**: Ожидает сигналы завершения и выполняет graceful shutdown.

**Поддерживаемые сигналы**: SIGINT (Ctrl+C), SIGTERM

**Graceful shutdown**:
- Прекращает принимать новые соединения
- Завершает активные запросы в рамках таймаута
- Закрывает idle соединения
- Уведомляет WebSocket соединения

---

## Примеры использования

### Создание поста

```go
// Создание входных данных
input := model.PostInput{
    Title:           "Заголовок моего поста",
    Content:         "Содержимое поста с подробным описанием",
    AuthorID:        userID,
    CommentsEnabled: true,
}

// Валидация
if err := input.Validate(); err != nil {
    return fmt.Errorf("ошибка валидации: %w", err)
}

// Создание поста через сервис
post, err := postService.CreatePost(ctx, input)
if err != nil {
    return fmt.Errorf("не удалось создать пост: %w", err)
}

fmt.Printf("Создан пост с ID: %s\n", post.ID)
```

### Создание комментария

```go
// Корневой комментарий
input := model.CommentInput{
    PostID:   postID,
    Content:  "Мой комментарий к посту",
    AuthorID: userID,
}

comment, err := commentService.CreateComment(ctx, input)
if err != nil {
    return err
}

// Ответ на комментарий
replyInput := model.CommentInput{
    PostID:   postID,
    ParentID: &comment.ID,
    Content:  "Мой ответ на комментарий",
    AuthorID: userID,
}

reply, err := commentService.CreateComment(ctx, replyInput)
```

### Подписка на события

```go
// Создание подписки
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

events, err := subscriptionService.Subscribe(ctx, postID)
if err != nil {
    return err
}

// Обработка событий
for event := range events {
    switch event.ActionType {
    case "CREATED":
        fmt.Printf("Новый комментарий: %s\n", event.Comment.Content)
    case "UPDATED":
        fmt.Printf("Обновлен комментарий: %s\n", event.Comment.Content)
    case "DELETED":
        fmt.Printf("Удален комментарий: %s\n", event.Comment.ID)
    }
}
```

### Построение дерева комментариев

```go
// Получение всех комментариев к посту
filter := model.CommentFilter{PostID: &postID}
connection, err := commentService.ListComments(ctx, filter, model.PaginationInput{})
if err != nil {
    return err
}

// Извлечение комментариев из connection
comments := make([]*model.Comment, len(connection.Edges))
for i, edge := range connection.Edges {
    comments[i] = edge.Node
}

// Построение дерева
tree := model.BuildCommentsTree(comments)

// Отображение дерева
for _, rootComment := range tree {
    printCommentTree(rootComment, 0)
}

func printCommentTree(comment *model.Comment, level int) {
    indent := strings.Repeat("  ", level)
    fmt.Printf("%s%s (depth: %d)\n", indent, comment.Content, comment.Depth)

    for _, child := range comment.Children {
        printCommentTree(child, level+1)
    }
}
```

---

## Соглашения по документации

### Стиль комментариев

1. **Структуры**: Описание назначения, основных полей и примеры использования
2. **Методы**: Назначение, параметры, возвращаемые значения, побочные эффекты, возможные ошибки
3. **Функции**: Подробное описание алгоритма, сложности, примеры использования
4. **Интерфейсы**: Общее назначение, ответственности реализации, примеры

### Форматирование

- Используются русские комментарии для понятности
- Структурированные списки для перечислений
- Примеры кода для демонстрации использования
- Указание возможных ошибок и их типов
- Описание алгоритмической сложности для функций

### Принципы

- **Полнота**: каждый публичный API документирован
- **Актуальность**: документация синхронизирована с кодом
- **Понятность**: объяснения доступны разработчикам разного уровня
- **Примеры**: практические демонстрации использования
