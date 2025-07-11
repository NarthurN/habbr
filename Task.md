Задача:

Реализовать систему для добавления и чтения постов и комментариев с использованием GraphQL, аналогичную комментариям к постам на популярных платформах, таких как Хабр или Reddit.

Характеристики системы постов:
•  Можно просмотреть список постов.
•  Можно просмотреть пост и комментарии под ним.
•  Пользователь, написавший пост, может запретить оставление комментариев к своему посту.

Характеристики системы комментариев к постам:
•  Комментарии организованы иерархически, позволяя вложенность без ограничений.
•  Длина текста комментария ограничена до, например, 2000 символов.
•  Система пагинации для получения списка комментариев.

(*) Дополнительные требования для реализации через GraphQL Subscriptions:
•  Комментарии к постам должны доставляться асинхронно, т.е. клиенты, подписанные на определенный пост, должны получать уведомления о новых комментариях без необходимости повторного запроса.

Требования к реализации:
•  Система должна быть написана на языке Go.
•  Использование Docker для распространения сервиса в виде Docker-образа.
•  Хранение данных может быть как в памяти (in-memory), так и в PostgreSQL. Выбор хранилища должен быть определяемым параметром при запуске сервиса.
•  Покрытие реализованного функционала unit-тестами.

Для PostgreSQL использовать библиотеку pgx для запросов.

Пример архитектуры проект
Архитектура проекта
Проект построен в соответствии с принципами чистой архитектуры и разделен на следующие слои:

1. API слой (Адаптеры)
Располагается в internal/api
Отвечает за обработку внешних запросов (gRPC)
Преобразует данные из внешнего формата (proto) во внутренние модели
Делегирует выполнение бизнес-логики в сервисный слой
2. Сервисный слой (Use Cases)
Располагается в internal/service
Содержит бизнес-логику приложения
Определяет интерфейсы репозиториев
Не зависит от внешних деталей (БД, протоколы и т.д.)
3. Репозиторный слой (Адаптеры)
Располагается в internal/repository
Отвечает за доступ к данным
Реализует интерфейсы, определенные в сервисном слое
Скрывает детали хранения данных от остальных слоев
Имеет собственные модели данных (internal/repository/model), отличные от доменных моделей
Использует конвертеры для преобразования между моделями репозитория и доменными моделями
4. Модели (Entities)
Располагаются в internal/model
Представляют основные бизнес-сущности
Не зависят от других слоев
5. Конвертеры
Располагаются в internal/converter и internal/repository/converter
Отвечают за преобразование данных между различными форматами
Обеспечивают изоляцию между слоями
Включают:
Конвертеры между Proto и доменными моделями (в internal/converter)
Конвертеры между доменными моделями и моделями репозитория (в internal/repository/converter)
Преимущества данной архитектуры
Разделение ответственностей - каждый слой имеет четко определенную ответственность
Изоляция зависимостей - зависимости направлены внутрь (к ядру приложения)
Тестируемость - слои можно тестировать изолированно
Гибкость - можно легко заменить конкретные реализации (например, базу данных)
Устойчивость к изменениям - изменения в одном слое минимально влияют на другие

.
├── README.md
├── Taskfile.yml
├── buf.work.yaml
├── go.work
├── go.work.sum
├── inventory
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   ├── go.sum
│   └── internal
│       ├── api
│       │   └── inventory
│       │       └── v1
│       │           ├── api.go
│       │           ├── get.go
│       │           └── list.go
│       ├── converter
│       │   └── part.go
│       ├── model
│       │   ├── errors.go
│       │   └── part.go
│       ├── repository
│       │   ├── converter
│       │   │   └── part.go
│       │   ├── mocks
│       │   │   └── mock_part_repository.go
│       │   ├── model
│       │   │   └── part.go
│       │   ├── part
│       │   │   ├── get.go
│       │   │   ├── init.go
│       │   │   ├── list.go
│       │   │   └── repository.go
│       │   └── repository.go
│       └── service
│           ├── mocks
│           │   └── mock_part_service.go
│           ├── part
│           │   ├── get.go
│           │   ├── get_test.go
│           │   ├── list.go
│           │   ├── list_test.go
│           │   ├── service.go
│           │   └── suite_test.go
│           └── service.go
├── order
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   ├── go.sum
│   └── internal
│       ├── api
│       │   └── order
│       │       └── v1
│       │           ├── api.go
│       │           ├── cancel.go
│       │           ├── create.go
│       │           ├── get.go
│       │           ├── new_order.go
│       │           └── pay.go
│       ├── client
│       │   ├── converter
│       │   │   └── part.go
│       │   └── grpc
│       │       ├── client.go
│       │       ├── inventory
│       │       │   └── v1
│       │       │       ├── client.go
│       │       │       └── list_parts.go
│       │       ├── mocks
│       │       │   ├── mock_inventory_client.go
│       │       │   └── mock_payment_client.go
│       │       └── payment
│       │           └── v1
│       │               ├── client.go
│       │               └── pay_order.go
│       ├── converter
│       │   └── order.go
│       ├── model
│       │   ├── error.go
│       │   ├── order.go
│       │   └── part.go
│       ├── repository
│       │   ├── converter
│       │   │   └── order.go
│       │   ├── mocks
│       │   │   └── mock_order_repository.go
│       │   ├── model
│       │   │   └── order.go
│       │   ├── order
│       │   │   ├── create.go
│       │   │   ├── get.go
│       │   │   ├── repository.go
│       │   │   └── update.go
│       │   └── repository.go
│       └── service
│           ├── mocks
│           │   └── mock_order_service.go
│           ├── order
│           │   ├── cancel.go
│           │   ├── cancel_test.go
│           │   ├── create.go
│           │   ├── create_test.go
│           │   ├── get.go
│           │   ├── get_test.go
│           │   ├── pay.go
│           │   ├── pay_test.go
│           │   ├── service.go
│           │   └── suite_test.go
│           └── service.go
├── package-lock.json
├── package.json
├── payment
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   ├── go.sum
│   └── internal
│       ├── api
│       │   └── payment
│       │       └── v1
│       │           ├── api.go
│       │           └── pay.go
│       ├── model
│       │   └── errors.go
│       └── service
│           ├── mocks
│           │   └── mock_payment_service.go
│           ├── payment
│           │   ├── pay.go
│           │   ├── pay_test.go
│           │   ├── service.go
│           │   └── suite_test.go
│           └── service.go
└── shared
    ├── api
    │   └── order
    │       └── v1
    │           ├── components
    │           │   ├── create_order_request.yaml
    │           │   ├── create_order_response.yaml
    │           │   ├── enums
    │           │   │   ├── order_status.yaml
    │           │   │   └── payment_method.yaml
    │           │   ├── errors
    │           │   │   ├── bad_gateway_error.yaml
    │           │   │   ├── bad_request_error.yaml
    │           │   │   ├── conflict_error.yaml
    │           │   │   ├── forbidden_error.yaml
    │           │   │   ├── generic_error.yaml
    │           │   │   ├── internal_server_error.yaml
    │           │   │   ├── not_found_error.yaml
    │           │   │   ├── rate_limit_error.yaml
    │           │   │   ├── service_unavailable_error.yaml
    │           │   │   ├── unauthorized_error.yaml
    │           │   │   └── validation_error.yaml
    │           │   ├── get_order_response.yaml
    │           │   ├── order_dto.yaml
    │           │   ├── pay_order_request.yaml
    │           │   └── pay_order_response.yaml
    │           ├── order.openapi.yaml
    │           ├── params
    │           │   └── order_uuid.yaml
    │           └── paths
    │               ├── order_by_uuid.yaml
    │               ├── order_cancel.yaml
    │               ├── order_pay.yaml
    │               └── orders.yaml
    ├── go.mod
    ├── go.sum
    ├── pkg
    │   ├── openapi
    │   │   └── order
    │   │       └── v1
    │   │           ├── oas_cfg_gen.go
    │   │           ├── oas_client_gen.go
    │   │           ├── oas_handlers_gen.go
    │   │           ├── oas_interfaces_gen.go
    │   │           ├── oas_json_gen.go
    │   │           ├── oas_labeler_gen.go
    │   │           ├── oas_middleware_gen.go
    │   │           ├── oas_operations_gen.go
    │   │           ├── oas_parameters_gen.go
    │   │           ├── oas_request_decoders_gen.go
    │   │           ├── oas_request_encoders_gen.go
    │   │           ├── oas_response_decoders_gen.go
    │   │           ├── oas_response_encoders_gen.go
    │   │           ├── oas_router_gen.go
    │   │           ├── oas_schemas_gen.go
    │   │           ├── oas_server_gen.go
    │   │           ├── oas_unimplemented_gen.go
    │   │           └── oas_validators_gen.go
    │   └── proto
    │       ├── inventory
    │       │   └── v1
    │       │       ├── inventory.pb.go
    │       │       └── inventory_grpc.pb.go
    │       └── payment
    │           └── v1
    │               ├── payment.pb.go
    │               └── payment_grpc.pb.go
    └── proto
        ├── buf.gen.yaml
        ├── buf.yaml
        ├── inventory
        │   └── v1
        │       └── inventory.proto
        └── payment
            └── v1
                └── payment.proto
