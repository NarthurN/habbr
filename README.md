# Habbr - GraphQL Posts & Comments API

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![GraphQL](https://img.shields.io/badge/GraphQL-API-pink.svg)](https://graphql.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-blue.svg)](https://docker.com/)

Современный высокопроизводительный GraphQL API для системы постов и комментариев, построенный с использованием Clean Architecture и лучших практик Go разработки.

## 🚀 Особенности

### Core Features
- **GraphQL API** с поддержкой queries, mutations и subscriptions
- **Иерархические комментарии** с неограниченной глубиной вложенности
- **Real-time уведомления** через WebSocket subscriptions
- **Cursor-based пагинация** для эффективной навигации
- **Полнотекстовый поиск** по постам и комментариям

### Architecture & Performance
- **Clean Architecture** с четким разделением слоев
- **Repository Pattern** с поддержкой PostgreSQL и in-memory storage
- **Dependency Injection** для тестируемости и модульности
- **Connection pooling** для оптимизации работы с БД
- **Индексы производительности** для быстрых запросов

### Developer Experience
- **Hot Reload** в режиме разработки
- **Docker контейнеризация** с multi-stage builds
- **Comprehensive тестирование** (unit, integration)
- **GraphQL Playground** для интерактивного тестирования
- **Structured logging** с уровнями и контекстом

### Production Ready
- **Graceful shutdown** с правильной обработкой сигналов
- **Health checks** для мониторинга
- **Metrics collection** для observability
- **Rate limiting** и защита от abuse
- **Security best practices** (validation, sanitization)

## 📋 Требования

- **Go 1.21+**
- **PostgreSQL 15+** (опционально)
- **Docker & Docker Compose** (рекомендуется)
- **Redis** (опционально, для кэширования)

## 🛠 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/NarthurN/habbr.git
cd habbr
```

### 2. Запуск с Docker (рекомендуется)

```bash
# Разработка с hot reload
make docker-dev

# Разработка с инструментами (pgAdmin, Redis Insight)
make docker-dev-tools

# Продакшн
make docker-prod
```

### 3. Локальная разработка

```bash
# Установка зависимостей и инструментов
make setup

# Запуск PostgreSQL
make db-up

# Запуск приложения
make dev
```

## 🔧 Конфигурация

### Переменные окружения

```bash
# Сервер
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
SERVER_SHUTDOWN_TIMEOUT=30s
SERVER_ENABLE_PLAYGROUND=true
SERVER_ENABLE_INTROSPECTION=true

# База данных
DATABASE_TYPE=postgres  # или memory
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=habbr
DATABASE_USER=habbr_user
DATABASE_PASSWORD=habbr_password
DATABASE_SSL_MODE=disable
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5
DATABASE_CONNECTION_MAX_LIFETIME=300s

# Логирование
LOGGER_LEVEL=info        # debug, info, warn, error
LOGGER_FORMAT=json       # json, console
LOGGER_ENABLE_CALLER=false
```

## 📚 API Документация

### GraphQL Schema

API предоставляет следующие основные типы:

#### Post
```graphql
type Post {
  id: ID!
  title: String!
  content: String!
  authorID: String!
  commentsEnabled: Boolean!
  createdAt: Time!
  updatedAt: Time!
  comments(first: Int, after: String, filter: CommentFilter): CommentConnection!
}
```

#### Comment
```graphql
type Comment {
  id: ID!
  postID: ID!
  parentID: ID
  content: String!
  authorID: String!
  depth: Int!
  createdAt: Time!
  updatedAt: Time!
  children(first: Int, after: String): CommentConnection!
}
```

### Примеры запросов

#### Получение постов с пагинацией
```graphql
query GetPosts($first: Int, $after: String) {
  posts(first: $first, after: $after) {
    edges {
      node {
        id
        title
        content
        authorID
        commentsEnabled
        createdAt
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

#### Создание поста
```graphql
mutation CreatePost($input: PostInput!) {
  createPost(input: $input) {
    success
    post {
      id
      title
      content
    }
    error
  }
}
```

#### Подписка на комментарии
```graphql
subscription CommentEvents($postID: ID!) {
  commentEvents(postID: $postID) {
    type
    comment {
      id
      content
      authorID
      depth
    }
    postID
  }
}
```

### Endpoints

- **GraphQL API**: `http://localhost:8080/query`
- **GraphQL Playground**: `http://localhost:8080/` (в режиме разработки)
- **Health Check**: `http://localhost:8080/health`
- **Metrics**: `http://localhost:8080/metrics`

## 🏗 Архитектура

Проект следует принципам Clean Architecture:

```
cmd/server/          # Application entry point
internal/
├── api/graphql/     # GraphQL layer (schemas, resolvers, converters)
├── service/         # Business logic layer
├── repository/      # Data access layer
├── model/          # Domain models
└── config/         # Configuration management
```

### Слои архитектуры

1. **API Layer** (`internal/api/graphql/`)
   - GraphQL схемы и резолверы
   - Конвертеры между GraphQL и domain типами
   - WebSocket subscriptions

2. **Service Layer** (`internal/service/`)
   - Бизнес-логика и use cases
   - Валидация и авторизация
   - Pub/Sub для real-time уведомлений

3. **Repository Layer** (`internal/repository/`)
   - Абстракция доступа к данным
   - Реализации для PostgreSQL и in-memory
   - Паттерн Repository с интерфейсами

4. **Domain Layer** (`internal/model/`)
   - Доменные модели и типы
   - Бизнес-правила и валидация
   - Агрегаты и value objects

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Только unit тесты
make test-unit

# Интеграционные тесты (требует PostgreSQL)
make test-integration

# Покрытие кода
make test-coverage
```

### Структура тестов

- **Unit тесты**: Тестирование отдельных компонентов с моками
- **Integration тесты**: Тестирование с реальной БД
- **Table-driven тесты**: Для полного покрытия edge cases
- **Моки**: Автогенерируемые с помощью testify

### Покрытие

Проект стремится к >80% покрытию кода:
- Domain models: 100%
- Services: >90%
- Repositories: >85%
- GraphQL converters: 100%

## 🐳 Docker

### Development

```bash
# Запуск в режиме разработки
./scripts/docker-dev.sh

# С инструментами разработки
./scripts/docker-dev.sh --with-tools
```

Доступные сервисы в режиме разработки:
- **API**: http://localhost:8080
- **pgAdmin**: http://localhost:5050 (admin@habbr.local / admin)
- **Redis Insight**: http://localhost:8001

### Production

```bash
# Запуск в продакшн режиме
./scripts/docker-prod.sh

# С логами
./scripts/docker-prod.sh --logs
```

### Multi-stage Build

Dockerfile включает несколько этапов:
- **builder**: Компиляция приложения
- **tester**: Запуск тестов
- **production**: Минимальный образ на scratch
- **development**: Образ с hot reload
- **debug**: Образ с инструментами отладки

## 🗄 База данных

### PostgreSQL Schema

База данных включает:
- **Таблицы**: `posts`, `comments`
- **Индексы**: Оптимизированные для иерархических запросов
- **Триггеры**: Автоматическое обновление timestamps и depth
- **Views**: Для агрегированной статистики
- **Functions**: Для сложных запросов и оптимизации

### Миграции

```bash
# Применение миграций
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

# Откат миграций
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" down 1
```

### Performance Features

- **Иерархические индексы** для быстрых комментариев
- **Частичные индексы** для специфических случаев
- **Полнотекстовый поиск** с GIN индексами
- **Материализованные представления** для аналитики
- **Рекурсивные CTE** для дерева комментариев

## 📊 Мониторинг и Observability

### Логирование

Структурированное логирование с помощью Zap:
- **Уровни**: DEBUG, INFO, WARN, ERROR
- **Контекст**: Request ID, User ID, операции
- **Форматы**: JSON (продакшн), Console (разработка)

### Метрики

- **HTTP метрики**: Latency, throughput, error rates
- **Database метрики**: Connection pool, query performance
- **Business метрики**: Posts created, comments count
- **Subscription метрики**: Active connections, messages sent

### Health Checks

- **Application health**: `/health`
- **Database connectivity**: Проверка пула соединений
- **External services**: Redis availability

## 🔒 Безопасность

### Input Validation
- **Санитизация**: Защита от XSS и injection
- **Length limits**: Контроль размера контента
- **UUID validation**: Проверка корректности идентификаторов

### Rate Limiting
- **Per-IP limits**: Защита от spam и abuse
- **Per-operation limits**: Разные лимиты для разных операций
- **Exponential backoff**: При превышении лимитов

### Database Security
- **Prepared statements**: Защита от SQL injection
- **Connection encryption**: SSL/TLS для продакшна
- **Least privilege**: Минимальные права доступа

## 🚀 Deployment

### Production Checklist

- [ ] Настроить переменные окружения
- [ ] Настроить SSL/TLS
- [ ] Настроить reverse proxy (nginx)
- [ ] Настроить мониторинг
- [ ] Настроить логирование
- [ ] Настроить бэкапы БД
- [ ] Настроить CI/CD pipeline

### Kubernetes

```yaml
# Пример deployment для Kubernetes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: habbr-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: habbr-api
  template:
    metadata:
      labels:
        app: habbr-api
    spec:
      containers:
      - name: habbr-api
        image: habbr/posts-comments-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_TYPE
          value: "postgres"
        - name: DATABASE_HOST
          value: "postgres-service"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 🤝 Участие в разработке

### Workflow

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Сделайте commit изменений (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Code Style

- Следуйте [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Используйте `make format` для форматирования
- Запускайте `make lint` перед commit
- Добавляйте тесты для нового функционала

### Commit Convention

```
type(scope): description

type: feat, fix, docs, style, refactor, test, chore
scope: api, service, repo, model, config, docker
```

## 📄 Лицензия

Этот проект лицензирован под MIT License. См. файл [LICENSE](LICENSE) для деталей.

## 🙏 Благодарности

- [gqlgen](https://github.com/99designs/gqlgen) - GraphQL генератор для Go
- [pgx](https://github.com/jackc/pgx) - PostgreSQL драйвер
- [testify](https://github.com/stretchr/testify) - Тестовый фреймворк
- [zap](https://github.com/uber-go/zap) - Структурированное логирование

## 📞 Поддержка

- **Issues**: [GitHub Issues](https://github.com/NarthurN/habbr/issues)
- **Discussions**: [GitHub Discussions](https://github.com/NarthurN/habbr/discussions)
- **Email**: support@habbr.dev

---

**Habbr** - Создавайте, комментируйте, взаимодействуйте! 🚀
