# Posts & Comments GraphQL API

Полноценный GraphQL API для системы постов и комментариев на Go с использованием чистой архитектуры.

## 🚀 Особенности

- **Clean Architecture** - четкое разделение на слои (API, Service, Repository, Domain)
- **GraphQL API** - мощный и гибкий API с поддержкой запросов, мутаций и подписок
- **Real-time уведомления** - WebSocket подписки на новые комментарии
- **Иерархические комментарии** - неограниченная вложенность комментариев
- **Пагинация** - cursor-based пагинация для масштабируемости
- **Переключаемое хранилище** - PostgreSQL или in-memory (конфигурируется)
- **Docker поддержка** - готовые образы и docker-compose
- **Observability** - структурированное логирование с Zap
- **Graceful shutdown** - корректное завершение работы
- **Health checks** - проверка состояния сервиса

## 📋 Функциональность

### Система постов
- ✅ Просмотр списка постов с пагинацией
- ✅ Просмотр отдельного поста с комментариями
- ✅ Создание, редактирование и удаление постов
- ✅ Возможность отключения комментариев автором поста

### Система комментариев
- ✅ Иерархические комментарии без ограничений по вложенности
- ✅ Ограничение длины текста комментария до 2000 символов
- ✅ Пагинация для получения списка комментариев
- ✅ CRUD операции для комментариев

### GraphQL Subscriptions
- ✅ Real-time уведомления о новых комментариях
- ✅ Подписка на комментарии к конкретному посту

## 🛠 Технологический стек

- **Go 1.23+** - основной язык
- **GraphQL** - gqlgen для генерации кода
- **PostgreSQL** - основная база данных (с pgx драйвером)
- **In-memory** - альтернативное хранилище для разработки
- **Docker** - контейнеризация
- **Zap** - структурированное логирование
- **UUID** - для идентификаторов
- **WebSocket** - для real-time подписок

## 📁 Структура проекта

```
posts-comments-graphql/
├── cmd/server/           # Точка входа приложения
├── internal/
│   ├── api/graphql/     # GraphQL резолверы и схемы (в graph/)
│   ├── service/         # Бизнес-логика
│   │   ├── post/        # Сервис постов
│   │   ├── comment/     # Сервис комментариев
│   │   └── subscription/ # Сервис подписок
│   ├── repository/      # Доступ к данным
│   │   ├── postgres/    # PostgreSQL реализация
│   │   ├── memory/      # In-memory реализация
│   │   ├── model/       # Модели репозитория
│   │   └── converter/   # Конвертеры
│   ├── model/           # Доменные сущности
│   ├── converter/       # Конвертеры между слоями
│   └── config/          # Конфигурация
├── graph/               # Сгенерированный GraphQL код
├── migrations/          # SQL миграции
├── docker/              # Docker файлы
├── tests/               # Тесты
└── Makefile            # Команды для разработки
```

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.23+
- Docker & Docker Compose (опционально)
- Make (опционально)

### Установка

1. **Клонируйте репозиторий:**
```bash
git clone <repo-url>
cd habbr
```

2. **Установите зависимости:**
```bash
make deps
# или
go mod download
```

3. **Сгенерируйте GraphQL код:**
```bash
make generate
# или
go run github.com/99designs/gqlgen generate
```

### Запуск

#### Локальный запуск (in-memory)
```bash
make run
# или
go run ./cmd/server
```

#### Docker (in-memory)
```bash
make docker-build
make docker-run
```

#### Docker Compose (PostgreSQL)
```bash
make docker-compose-up
```

#### Docker Compose (in-memory для разработки)
```bash
docker-compose --profile memory up
```

## 🔧 Конфигурация

Настройка через переменные окружения:

### Сервер
- `SERVER_HOST` - хост сервера (по умолчанию: localhost)
- `SERVER_PORT` - порт сервера (по умолчанию: 8080)
- `SERVER_ENABLE_PLAYGROUND` - включить GraphQL Playground (по умолчанию: true)
- `SERVER_ENABLE_INTROSPECTION` - включить интроспекцию (по умолчанию: true)

### База данных
- `DATABASE_TYPE` - тип БД: "postgres" или "memory" (по умолчанию: memory)
- `DATABASE_HOST` - хост PostgreSQL (по умолчанию: localhost)
- `DATABASE_PORT` - порт PostgreSQL (по умолчанию: 5432)
- `DATABASE_NAME` - имя БД (по умолчанию: habbr)
- `DATABASE_USER` - пользователь БД (по умолчанию: habbr)
- `DATABASE_PASSWORD` - пароль БД (по умолчанию: password)

### Логирование
- `LOGGER_LEVEL` - уровень логов: debug, info, warn, error (по умолчанию: info)
- `LOGGER_FORMAT` - формат логов: json, console (по умолчанию: json)
- `LOGGER_ENABLE_CALLER` - включить информацию о вызывающем коде (по умолчанию: true)

## 📝 GraphQL API

### Endpoints

- **GraphQL API**: `http://localhost:8080/query`
- **GraphQL Playground**: `http://localhost:8080/`
- **Health Check**: `http://localhost:8080/health`

### Примеры запросов

#### Создание поста
```graphql
mutation {
  createPost(input: {
    title: "Мой первый пост"
    content: "Содержание поста..."
    authorId: "550e8400-e29b-41d4-a716-446655440000"
    commentsEnabled: true
  }) {
    id
    title
    createdAt
  }
}
```

#### Получение постов с пагинацией
```graphql
query {
  posts(first: 10) {
    edges {
      node {
        id
        title
        content
        commentCount
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

#### Создание комментария
```graphql
mutation {
  createComment(input: {
    postId: "550e8400-e29b-41d4-a716-446655440001"
    content: "Отличный пост!"
    authorId: "550e8400-e29b-41d4-a716-446655440002"
  }) {
    id
    content
    depth
    createdAt
  }
}
```

#### Получение дерева комментариев
```graphql
query {
  commentsTree(postId: "550e8400-e29b-41d4-a716-446655440001") {
    id
    content
    depth
    children {
      id
      content
      depth
    }
  }
}
```

#### Подписка на комментарии
```graphql
subscription {
  commentUpdates(postId: "550e8400-e29b-41d4-a716-446655440001") {
    postId
    actionType
    comment {
      id
      content
      authorId
    }
  }
}
```

## 🧪 Тестирование

```bash
# Запуск тестов
make test

# Тесты с покрытием
make test-coverage

# Только unit тесты
go test ./internal/...
```

## 🔍 Разработка

### Полезные команды

```bash
# Генерация GraphQL кода
make generate

# Форматирование кода
make format

# Линтер
make lint

# Сборка
make build

# Очистка
make clean

# Установка инструментов разработки
make install-tools

# Полная настройка окружения
make setup
```

### Добавление новых типов GraphQL

1. Обновите схему в `graph/schema.graphqls`
2. Запустите `make generate`
3. Реализуйте резолверы в `graph/schema.resolvers.go`

### Добавление новых репозиториев

1. Создайте интерфейс в `internal/repository/interfaces.go`
2. Реализуйте для memory в `internal/repository/memory/`
3. Реализуйте для postgres в `internal/repository/postgres/`

## 🐳 Docker

### Локальная сборка
```bash
docker build -t habbr/posts-comments-api:latest .
```

### Многоэтапная сборка
Dockerfile использует многоэтапную сборку для минимизации размера образа:
- **Builder stage**: Go 1.23.4 Alpine для сборки
- **Final stage**: scratch образ с только бинарным файлом

### Docker Compose профили
- **По умолчанию**: app + postgres
- **Memory профиль**: app-memory (in-memory БД)

## 📊 Архитектура

### Слои

1. **API слой** (`graph/`) - GraphQL резолверы
2. **Service слой** (`internal/service/`) - бизнес-логика
3. **Repository слой** (`internal/repository/`) - доступ к данным
4. **Domain слой** (`internal/model/`) - доменные модели

### Зависимости

```
API Layer → Service Layer → Repository Layer
     ↓           ↓              ↓
  GraphQL    Business      Data Access
  Resolvers   Logic        (Postgres/Memory)
```

### Принципы

- **Dependency Injection** - внедрение зависимостей через интерфейсы
- **Interface Segregation** - мелкие, специфичные интерфейсы
- **Single Responsibility** - каждый компонент имеет одну ответственность
- **Open/Closed** - открыт для расширения, закрыт для изменения

## 🔒 Безопасность

- Валидация входных данных на всех уровнях
- Контроль длины комментариев (макс. 2000 символов)
- Проверка прав доступа (автор может редактировать/удалять)
- Graceful обработка ошибок

## 📈 Производительность

- Cursor-based пагинация для больших наборов данных
- Кэширование GraphQL запросов
- Эффективные индексы в PostgreSQL
- Минимальные аллокации в Go коде

## 🤝 Участие в разработке

1. Fork репозитория
2. Создайте feature branch: `git checkout -b feature/amazing-feature`
3. Зафиксируйте изменения: `git commit -m 'Add amazing feature'`
4. Push в branch: `git push origin feature/amazing-feature`
5. Создайте Pull Request

## 📄 Лицензия

MIT License - см. файл [LICENSE](LICENSE)

## 🆘 Поддержка

Если у вас есть вопросы или проблемы:

1. Проверьте [Issues](../../issues)
2. Создайте новый Issue с подробным описанием
3. Используйте соответствующие теги (bug, feature, question)

---

**Автор**: [NarthurN](https://github.com/NarthurN)
**Версия**: 1.0.0
