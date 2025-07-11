# Contributing Guide - Habbr

Спасибо за ваш интерес к участию в разработке Habbr! Это руководство поможет вам начать работу над проектом.

## 📋 Оглавление

- [Кодекс поведения](#кодекс-поведения)
- [Как внести вклад](#как-внести-вклад)
- [Настройка окружения разработки](#настройка-окружения-разработки)
- [Процесс разработки](#процесс-разработки)
- [Стандарты кода](#стандарты-кода)
- [Тестирование](#тестирование)
- [Документация](#документация)
- [Pull Request процесс](#pull-request-процесс)

## Кодекс поведения

Проект придерживается принципов открытости и инклюзивности. Все участники должны:

- Быть уважительными к другим участникам
- Принимать конструктивную критику
- Фокусироваться на том, что лучше для сообщества
- Проявлять эмпатию к другим участникам сообщества

## Как внести вклад

### Типы вклада

Мы приветствуем различные типы вклада:

- 🐛 **Сообщения об ошибках**
- 💡 **Предложения новых функций**
- 📝 **Улучшения документации**
- 🔧 **Исправления кода**
- ✨ **Новые функции**
- 🧪 **Написание тестов**
- 🔍 **Код ревью**

### Прежде чем начать

1. Проверьте [существующие Issues](../../issues)
2. Ознакомьтесь с [архитектурой проекта](ARCHITECTURE.md)
3. Прочитайте это руководство полностью

## Настройка окружения разработки

### Требования

- Go 1.21+
- Docker & Docker Compose
- Git
- Make (опционально)

### Установка

1. **Клонируйте репозиторий:**
   ```bash
   git clone https://github.com/NarthurN/habbr.git
   cd habbr
   ```

2. **Настройте окружение:**
   ```bash
   make setup
   ```

3. **Запустите в режиме разработки:**
   ```bash
   make docker-dev-tools
   ```

4. **Проверьте, что все работает:**
   ```bash
   # Запуск тестов
   make test

   # Проверка линтера
   make lint

   # Проверка форматирования
   make format
   ```

### Структура проекта

```
habbr/
├── cmd/server/          # Точка входа приложения
├── internal/
│   ├── api/graphql/    # GraphQL API слой
│   ├── service/        # Бизнес-логика
│   ├── repository/     # Доступ к данным
│   ├── model/          # Доменные модели
│   └── config/         # Конфигурация
├── migrations/         # SQL миграции
├── scripts/           # Вспомогательные скрипты
├── docs/              # Документация
└── tests/             # Интеграционные тесты
```

## Процесс разработки

### Workflow

1. **Создайте Issue** (если его нет)
2. **Создайте feature branch** из `main`
3. **Разработайте функцию** с тестами
4. **Проверьте код** линтером и тестами
5. **Создайте Pull Request**
6. **Пройдите код ревью**
7. **Merge** после одобрения

### Именование веток

```bash
# Новые функции
feature/add-user-authentication
feature/implement-rate-limiting

# Исправления ошибок
fix/comment-depth-validation
fix/memory-leak-in-subscriptions

# Документация
docs/update-api-examples
docs/add-deployment-guide

# Рефакторинг
refactor/extract-validation-service
refactor/improve-error-handling

# Chore (технические задачи)
chore/update-dependencies
chore/setup-ci-pipeline
```

### Commit сообщения

Используем [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

body (опционально)

footer (опционально)
```

#### Типы:

- `feat`: новая функция
- `fix`: исправление ошибки
- `docs`: изменения документации
- `style`: форматирование кода
- `refactor`: рефакторинг без изменения функциональности
- `test`: добавление или изменение тестов
- `chore`: изменения в инфраструктуре

#### Примеры:

```bash
feat(api): add rate limiting middleware
fix(repository): handle concurrent access to in-memory storage
docs(readme): update installation instructions
test(service): add unit tests for comment validation
refactor(resolver): extract common error handling
```

## Стандарты кода

### Go Style Guide

Следуем [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments):

#### Именование

```go
// Good: используйте краткие, описательные имена
func GetPost(id uuid.UUID) (*Post, error)
func ValidateInput(input *PostInput) error

// Bad: избегайте сокращений и неясных имен
func GetP(i uuid.UUID) (*P, error)
func ValInp(inp *PI) error
```

#### Обработка ошибок

```go
// Good: всегда обрабатывайте ошибки
post, err := repo.GetPost(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get post %s: %w", id, err)
}

// Good: используйте типизированные ошибки
if errors.Is(err, ErrPostNotFound) {
    return nil, ErrPostNotFound
}

// Bad: игнорирование ошибок
post, _ := repo.GetPost(ctx, id)
```

#### Интерфейсы

```go
// Good: маленькие, специфические интерфейсы
type PostRepository interface {
    GetPost(ctx context.Context, id uuid.UUID) (*Post, error)
    CreatePost(ctx context.Context, input *PostInput) (*Post, error)
}

// Good: принимайте интерфейсы, возвращайте структуры
func NewPostService(repo PostRepository) *PostService {
    return &PostService{repo: repo}
}
```

#### Структуры

```go
// Good: группируйте связанные поля
type Post struct {
    // Identity
    ID       uuid.UUID
    AuthorID uuid.UUID

    // Content
    Title   string
    Content string

    // Settings
    CommentsEnabled bool

    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### GraphQL Style Guide

#### Именование

```graphql
# Good: используйте PascalCase для типов
type Post {
  id: ID!
  title: String!
}

# Good: используйте camelCase для полей
type Query {
  posts(first: Int, after: String): PostConnection!
  searchPosts(query: String!): [Post!]!
}
```

#### Пагинация

```graphql
# Good: используйте Relay Cursor Connections
type PostConnection {
  edges: [PostEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PostEdge {
  node: Post!
  cursor: String!
}
```

### Форматирование

```bash
# Автоматическое форматирование
make format

# Или вручную
go fmt ./...
goimports -w .
```

### Линтинг

```bash
# Запуск линтера
make lint

# Или вручную
golangci-lint run
```

## Тестирование

### Принципы тестирования

1. **Пирамида тестов**: больше unit тестов, меньше integration тестов
2. **Table-driven тесты**: для покрытия множественных сценариев
3. **Моки**: для изоляции unit тестов
4. **Четкие имена**: описывающие что тестируется

### Unit тесты

```go
func TestPostService_CreatePost(t *testing.T) {
    tests := []struct {
        name        string
        input       *model.PostInput
        setupMock   func(*mocks.PostRepository)
        expected    *model.Post
        expectError bool
        errorMsg    string
    }{
        {
            name: "successful creation",
            input: &model.PostInput{
                Title:   "Test Post",
                Content: "Test Content",
                AuthorID: uuid.New(),
            },
            setupMock: func(mock *mocks.PostRepository) {
                mock.On("CreatePost", mock.Anything, mock.Anything).
                    Return(&model.Post{ID: uuid.New()}, nil)
            },
            expectError: false,
        },
        {
            name: "validation error",
            input: &model.PostInput{
                Title: "", // Invalid empty title
            },
            expectError: true,
            errorMsg:    "title cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := &mocks.PostRepository{}
            if tt.setupMock != nil {
                tt.setupMock(mockRepo)
            }

            service := service.NewPostService(mockRepo, nil, zap.NewNop())

            // Act
            result, err := service.CreatePost(context.Background(), tt.input)

            // Assert
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Integration тесты

```go
//go:build integration

func TestPostRepository_Integration(t *testing.T) {
    // Настройка тестовой БД
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := postgres.NewPostRepository(db)

    t.Run("create and retrieve post", func(t *testing.T) {
        input := &model.PostInput{
            Title:    "Integration Test Post",
            Content:  "Content for integration test",
            AuthorID: uuid.New(),
        }

        // Create
        created, err := repo.CreatePost(context.Background(), input)
        require.NoError(t, err)
        require.NotNil(t, created)
        assert.Equal(t, input.Title, created.Title)

        // Retrieve
        retrieved, err := repo.GetPost(context.Background(), created.ID)
        require.NoError(t, err)
        assert.Equal(t, created.ID, retrieved.ID)
        assert.Equal(t, created.Title, retrieved.Title)
    })
}
```

### Запуск тестов

```bash
# Все тесты
make test

# Только unit тесты
make test-unit

# Только integration тесты (требует БД)
make test-integration

# С покрытием
make test-coverage

# Конкретный пакет
go test -v ./internal/service/...

# Конкретный тест
go test -v -run TestPostService_CreatePost ./internal/service/
```

### Покрытие

Стремимся к следующему покрытию:
- **Domain models**: 100%
- **Services**: >90%
- **Repositories**: >85%
- **API converters**: 100%
- **Общее покрытие**: >80%

## Документация

### Требования к документации

1. **GoDoc комментарии** для всех экспортируемых функций
2. **README** обновления при изменении API
3. **Примеры использования** для новых функций
4. **Архитектурная документация** для крупных изменений

### GoDoc стиль

```go
// PostService предоставляет бизнес-логику для работы с постами.
// Сервис обеспечивает валидацию, авторизацию и координацию между репозиториями.
type PostService struct {
    postRepo    repository.PostRepository
    commentRepo repository.CommentRepository
    logger      *zap.Logger
}

// CreatePost создает новый пост с валидацией входных данных.
// Возвращает созданный пост или ошибку валидации/создания.
//
// Пример использования:
//
//	input := &model.PostInput{
//		Title:   "My Post",
//		Content: "Post content",
//		AuthorID: authorID,
//	}
//	post, err := service.CreatePost(ctx, input)
//	if err != nil {
//		// handle error
//	}
func (s *PostService) CreatePost(ctx context.Context, input *model.PostInput) (*model.Post, error) {
    // implementation
}
```

## Pull Request процесс

### Подготовка PR

1. **Убедитесь, что ваша ветка актуальна:**
   ```bash
   git checkout main
   git pull origin main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Запустите проверки:**
   ```bash
   make test
   make lint
   make format
   ```

3. **Убедитесь в качестве коммитов:**
   ```bash
   git log --oneline
   # При необходимости сделайте rebase для чистой истории
   ```

### Описание PR

Используйте следующий шаблон:

```markdown
## Описание

Краткое описание изменений и причины их внесения.

## Тип изменений

- [ ] Исправление ошибки (non-breaking change)
- [ ] Новая функция (non-breaking change)
- [ ] Breaking change (исправление или функция, которая изменяет существующую функциональность)
- [ ] Изменения документации

## Как протестировано

- [ ] Unit тесты проходят
- [ ] Integration тесты проходят (если применимо)
- [ ] Добавлены новые тесты
- [ ] Ручное тестирование выполнено

## Чеклист

- [ ] Код следует стандартам проекта
- [ ] Саморевью кода выполнено
- [ ] Код прокомментирован в сложных местах
- [ ] Документация обновлена
- [ ] Нет конфликтов с main веткой
- [ ] Функциональность протестирована

## Связанные Issues

Fixes #123
Related to #456

## Скриншоты (если применимо)

[Добавьте скриншоты для UI изменений]
```

### Процесс ревью

1. **Автоматические проверки** должны пройти
2. **Минимум один approve** от мейнтейнера
3. **Решение всех комментариев** ревьюера
4. **Обновление документации** при необходимости

### После мержа

1. **Удалите feature ветку**
2. **Обновите локальную main ветку**
3. **Закройте связанные Issues**

## Частые вопросы

### Q: Как добавить новый GraphQL тип?

1. Добавьте тип в схему (`internal/api/graphql/schema/`)
2. Запустите `make generate`
3. Реализуйте резолверы
4. Добавьте конвертеры
5. Напишите тесты

### Q: Как добавить новое поле в существующий тип?

1. Обновите доменную модель (`internal/model/`)
2. Обновите GraphQL схему
3. Запустите `make generate`
4. Обновите конвертеры
5. Обновите репозитории (если нужно)
6. Добавьте миграцию БД (если нужно)

### Q: Как работает система пагинации?

Используем Relay Cursor Connections. См. примеры в `internal/repository/` и документацию GraphQL.

### Q: Как добавить новый тест?

1. Unit тесты: в том же пакете с суффиксом `_test.go`
2. Integration тесты: с build tag `//go:build integration`
3. Следуйте существующим паттернам

## Получение помощи

- **Issues**: [GitHub Issues](../../issues)
- **Discussions**: [GitHub Discussions](../../discussions)
- **Документация**: [docs/](../docs/)

Спасибо за ваш вклад в Habbr! 🚀
