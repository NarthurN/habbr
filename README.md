# Habbr - GraphQL API для постов и комментариев

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![GraphQL](https://img.shields.io/badge/GraphQL-API-pink.svg)](https://graphql.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-blue.svg)](https://docker.com/)

## 📝 Описание проекта

**Habbr** — это современная система для создания и управления постами с иерархическими комментариями, построенная на GraphQL API. Проект реализует функциональность, аналогичную комментариям на популярных платформах, таких как Хабр или Reddit.

### Зачем создан проект

Проект был создан для изучения и демонстрации современных подходов к разработке на Go:
- **GraphQL** как альтернатива REST API с возможностями real-time подписок
- **Clean Architecture** для создания поддерживаемого и тестируемого кода
- **Микросервисные паттерны** с dependency injection и repository pattern
- **Production-ready решения** с Docker, мониторингом и observability

### Решаемые задачи

- **Иерархические комментарии** с неограниченной глубиной вложенности
- **Real-time уведомления** о новых комментариях через WebSocket subscriptions
- **Эффективная пагинация** с cursor-based подходом для больших объемов данных
- **Гибкое хранение данных** с поддержкой PostgreSQL и in-memory режимов
- **Высокая производительность** с оптимизированными запросами и индексами

### Технологический стек

- **Backend**: Go 1.23+, GraphQL (gqlgen)
- **База данных**: PostgreSQL 15+ с pgx драйвером
- **Кэширование**: Redis 7+
- **Контейнеризация**: Docker & Docker Compose
- **Тестирование**: testify, table-driven тесты
- **Логирование**: Zap structured logging
- **Архитектура**: Clean Architecture, Repository Pattern

## 🚀 Быстрый запуск

### Системные требования

- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Git** для клонирования репозитория

### Запуск через Docker Compose

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/NarthurN/habbr.git
cd habbr
```

2. **Запустите все сервисы:**
```bash
# Базовый запуск (API + PostgreSQL + Redis)
docker compose up -d

# Запуск с инструментами разработки (+ pgAdmin + Redis Insight)
docker compose --profile tools up -d
```

3. **Проверьте статус сервисов:**
```bash
docker compose ps
```

### Проверка работоспособности

После успешного запуска вы можете проверить работу системы:

#### 🎯 Основные endpoints:
- **GraphQL Playground**: http://localhost:8080/ - интерактивная среда для тестирования API
- **GraphQL API**: http://localhost:8080/query - основной endpoint для запросов
- **Health Check**: http://localhost:8080/health - проверка состояния сервиса

#### 🛠 Инструменты разработки (при запуске с `--profile tools`):
- **pgAdmin**: http://localhost:5050 - веб-интерфейс для PostgreSQL
  - Email: `admin@habbr.local`
  - Password: `admin`
- **Redis Insight**: http://localhost:8001 - интерфейс для мониторинга Redis

#### 🧪 Быстрый тест API

Откройте GraphQL Playground (http://localhost:8080/) и выполните тестовые запросы:

**1. Создание поста:**
```graphql
mutation {
  createPost(input: {
    title: "Мой первый пост"
    content: "Это содержимое моего первого поста в Habbr"
    authorID: "user-123"
    commentsEnabled: true
  }) {
    success
    post {
      id
      title
      createdAt
    }
    error
  }
}
```

**2. Получение списка постов:**
```graphql
query {
  posts(first: 10) {
    edges {
      node {
        id
        title
        content
        authorID
        commentsEnabled
        createdAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

**3. Создание комментария:**
```graphql
mutation {
  createComment(input: {
    postID: "YOUR_POST_ID"  # Замените на ID созданного поста
    content: "Отличный пост!"
    authorID: "user-456"
  }) {
    success
    comment {
      id
      content
      depth
      createdAt
    }
    error
  }
}
```

### Остановка сервисов

```bash
# Остановка всех сервисов
docker compose down

# Остановка с удалением данных
docker compose down -v
```

## 🏗 Архитектура проекта

Проект следует принципам **Clean Architecture** с четким разделением ответственности:

```
cmd/server/          # Точка входа приложения
internal/
├── api/graphql/     # GraphQL слой (схемы, резолверы, конвертеры)
│   ├── schema/      # GraphQL схемы
│   ├── resolver/    # Резолверы для queries, mutations, subscriptions
│   └── converter/   # Конвертеры между GraphQL и domain типами
├── service/         # Бизнес-логика
│   ├── post/        # Сервис работы с постами
│   ├── comment/     # Сервис работы с комментариями
│   └── subscription/ # Сервис real-time подписок
├── repository/      # Слой доступа к данным
│   ├── postgres/    # PostgreSQL реализация
│   ├── memory/      # In-memory реализация
│   └── model/       # Модели репозитория
├── model/          # Доменные модели
└── config/         # Управление конфигурацией
```

### Ключевые принципы:

- **Dependency Injection** через интерфейсы
- **Repository Pattern** для абстракции хранения данных
- **Publisher-Subscriber** для real-time уведомлений
- **Cursor-based pagination** для эффективной навигации
- **Graceful shutdown** с корректной обработкой сигналов

## 🔧 Конфигурация

### Переменные окружения

Основные настройки можно изменить через переменные окружения:

```bash
# Тип хранилища данных
DATABASE_TYPE=postgres          # или "memory" для in-memory режима

# Настройки PostgreSQL
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=habbr
DATABASE_USER=habbr_user
DATABASE_PASSWORD=habbr_password

# Настройки сервера
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_ENABLE_PLAYGROUND=true   # Включить GraphQL Playground

# Логирование
LOGGER_LEVEL=info               # debug, info, warn, error
LOGGER_FORMAT=json              # json или console
```

### Запуск с in-memory хранилищем

Для быстрого тестирования без PostgreSQL:

```bash
# Создайте файл .env
echo "DATABASE_TYPE=memory" > .env

# Запустите только API сервис
docker compose up habbr-api
```

## 🧪 Тестирование

### Запуск тестов

```bash
# Unit тесты
docker compose exec habbr-api go test ./internal/...

# Интеграционные тесты (требует PostgreSQL)
docker compose exec habbr-api go test -tags=integration ./...

# Покрытие кода
docker compose exec habbr-api go test -cover ./...
```

### Структура тестов

- **Unit тесты**: Изолированное тестирование компонентов с моками
- **Integration тесты**: Тестирование с реальной базой данных
- **Table-driven тесты**: Полное покрытие edge cases
- **GraphQL тесты**: Тестирование API через GraphQL queries

## 📊 Мониторинг и отладка

### Логи приложения

```bash
# Просмотр логов API
docker compose logs -f habbr-api

# Логи PostgreSQL
docker compose logs -f postgres

# Логи всех сервисов
docker compose logs -f
```

### Метрики производительности

- **Health Check**: http://localhost:8080/health
- **Database connections**: Мониторинг через pgAdmin
- **Redis metrics**: Мониторинг через Redis Insight

### Отладка проблем

1. **Проверьте статус сервисов**: `docker compose ps`
2. **Просмотрите логи**: `docker compose logs habbr-api`
3. **Проверьте health check**: `curl http://localhost:8080/health`
4. **Убедитесь в доступности БД**: через pgAdmin или прямое подключение

## 🚀 Планы по доработке

### Краткосрочные планы (1-2 месяца):

1. **Аутентификация и авторизация**
   - Интеграция с JWT токенами
   - Ролевая модель доступа
   - Защита GraphQL endpoints

2. **Оптимизация производительности**
   - Добавление Redis кэширования для частых запросов
   - Реализация DataLoader для решения N+1 проблемы в GraphQL
   - Добавление индексов для полнотекстового поиска

3. **Расширение API**
   - Поддержка файловых вложений в постах и комментариях
   - Система лайков и рейтингов
   - Уведомления пользователей

### Долгосрочные планы (3-6 месяцев):

1. **Микросервисная архитектура**
   - Выделение сервиса уведомлений
   - Сервис управления пользователями
   - Event-driven архитектура с Apache Kafka

2. **Масштабирование**
   - Горизонтальное масштабирование API
   - Шардинг базы данных
   - CDN для статических ресурсов

3. **DevOps и мониторинг**
   - CI/CD pipeline с GitHub Actions
   - Kubernetes deployment
   - Prometheus + Grafana для мониторинга
   - Distributed tracing с Jaeger

## 📚 Дополнительная информация

### GraphQL Schema

Полная документация API доступна в GraphQL Playground. Основные типы:

- **Post**: Представляет пост с заголовком, содержимым и настройками комментариев
- **Comment**: Иерархический комментарий с поддержкой вложенности
- **Connections**: Cursor-based пагинация для списков
- **Subscriptions**: Real-time уведомления о новых комментариях

### Примеры использования

В директории `docs/` находятся:
- **API_EXAMPLES.md**: Подробные примеры GraphQL запросов
- **ARCHITECTURE.md**: Детальное описание архитектуры
- **CONTRIBUTING.md**: Руководство для разработчиков

### База данных

- **Миграции**: Автоматически применяются при запуске
- **Индексы**: Оптимизированы для иерархических запросов
- **Constraints**: Обеспечивают целостность данных
- **Triggers**: Автоматическое обновление метаданных

### Performance

Система оптимизирована для работы с большими объемами данных:
- **Эффективные SQL запросы** с минимальным количеством обращений к БД
- **Индексы производительности** для быстрого поиска комментариев
- **Connection pooling** для оптимального использования ресурсов
- **Cursor-based пагинация** для стабильной навигации

---

**Habbr** - создавайте контент, обсуждайте идеи, строите сообщества! 🚀
