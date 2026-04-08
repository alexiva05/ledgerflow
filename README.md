# LedgerFlow

Banking core — учебный production-ready микросервисный бэкенд на Go.

Реализует двойную бухгалтерскую запись (double-entry ledger), идемпотентные переводы, event-driven архитектуру и audit log.

---

## Архитектурные решения

### Double-entry ledger
Любая финансовая операция создаёт ровно две записи в `journal_entries` — debit и credit.
Инвариант: `SUM(debit) = SUM(credit)` по каждому `transaction_id`. Нарушение — баг, не бизнес-сценарий.

### Idempotency
Каждый `POST /transfers` принимает заголовок `Idempotency-Key` (UUID от клиента).
Повторный запрос с тем же ключом возвращает кешированный результат без повторного выполнения.

### Outbox pattern
Событие пишется в таблицу `outbox` **в той же PostgreSQL-транзакции**, что и бизнес-данные.
Отдельный воркер читает `outbox WHERE sent_at IS NULL` и публикует в Kafka.
Гарантия: at-least-once delivery.

### SAGA
Переводы между счетами разных сервисов — оркестрируемая SAGA с компенсирующими шагами.
При сбое на любом шаге запускается `compensate()`.

---

## Сервисы

| Сервис | Назначение | Статус |
|---|---|---|
| `account` | Double-entry ledger, балансы | Готово |
| `transaction` | Переводы, idempotency, outbox, gRPC | Готово |
| `fraud` | Velocity check через Redis, Kafka consumer | В разработке |
| `auth` | JWT, refresh tokens, RBAC | Planned |
| `notification` | Email / push / webhook | Planned |
| `reporting` | Выписки, аналитика | Planned |
| `audit` | Append-only signed log | Planned |

---

## Стек

| Категория | Технология |
|---|---|
| HTTP API | Gin |
| Internal RPC | gRPC + protobuf |
| База данных | PostgreSQL 16 + pgx v5 |
| Миграции | goose |
| SQL codegen | sqlc |
| Кэш / locks | Redis 7 + go-redis |
| Events | Kafka + franz-go |
| Аналитика | ClickHouse |
| Числа | shopspring/decimal (никогда float64) |
| DI | wire |
| Config | viper + env (12-factor) |
| Secrets | HashiCorp Vault |
| Tracing | OpenTelemetry + Jaeger |
| Metrics | Prometheus + Grafana |
| Logs | uber-go/zap |
| Тесты | testify + testcontainers-go |
| Lint | golangci-lint |

---

## Структура монорепо

```
ledgerflow/
├── services/
│   ├── account/           # double-entry ledger
│   │   ├── cmd/server/    # точка входа
│   │   ├── internal/
│   │   │   ├── domain/    # Account, JournalEntry, интерфейсы
│   │   │   ├── app/       # use cases
│   │   │   ├── infra/     # postgres, kafka, redis
│   │   │   └── transport/ # HTTP handlers
│   │   └── migrations/
│   └── transaction/       # переводы, idempotency, outbox
│       ├── internal/
│       │   ├── domain/    # Transaction, репозитории
│       │   └── ...
│       └── migrations/
├── pkg/
│   ├── logger/            # zap wrapper, FromContext/WithContext
│   ├── config/            # viper + env validation
│   ├── errors/            # доменные типы, маппинг в HTTP-статусы
│   └── tracing/           # OTel init, StartSpan, Gin middleware
├── proto/                 # .proto контракты (shared)
├── infra/                 # Docker Compose, Helm
└── go.work
```

---

## Kafka топики

| Топик | Producer | Consumers |
|---|---|---|
| `transaction.created` | transaction | fraud, audit, reporting |
| `transaction.completed` | transaction | notification, audit |
| `transaction.failed` | transaction | notification, audit |
| `balance.updated` | account | audit, reporting |
| `fraud.alert` | fraud | transaction, notification, audit |

Партиционирование по `account_id` — гарантирует порядок событий для одного счёта.

---

## Быстрый старт

```bash
# Поднять инфраструктуру
docker compose up -d

# Применить миграции
make migrate-up

# Запустить сервис
go run ./services/account/cmd/server
go run ./services/transaction/cmd/server

# Тесты (unit)
go test ./...

# Интеграционные тесты (требуют Docker)
go test ./services/account/... -tags integration

# Линтер
make lint
```

---

## Переменные окружения

```env
# account service
DATABASE_URL=postgres://ledger:secret@localhost:5432/account_db
REDIS_URL=redis://localhost:6379
LOG_LEVEL=info
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# transaction service
DATABASE_URL=postgres://ledger:secret@localhost:5432/transaction_db
KAFKA_BROKERS=localhost:9092
ACCOUNT_SERVICE_ADDR=localhost:50051
VAULT_ADDR=http://localhost:8200
```

---

## Соглашения по коду

**Деньги** — только `decimal.Decimal`, колонки `NUMERIC(20,4)`. Float64 — баг.

**Context** — первый аргумент везде, никогда не хранить в структурах.

**Ошибки** — типизированные доменные ошибки, HTTP-слой маппит в статус-коды.

**Логи** — только `zap`, структурированные поля, никаких `fmt.Println`.

**Транзакции PostgreSQL** — явно: `begin → work → commit`, `defer tx.Rollback` как safety net.

**Интерфейсы** — везде где есть I/O. Тестируемость — не опция.

---

## Роадмап

- [x] Фаза 1 — `pkg/` (logger, config, errors, tracing)
- [x] Фаза 2 — Account service (domain, storage, HTTP, тесты)
- [x] Фаза 3 — Transaction service (idempotency, outbox, gRPC, тесты)
- [ ] Фаза 4 — Event consumers (fraud — в разработке, audit, notification)
- [ ] Фаза 5 — Observability (Prometheus, OTel tracing, Grafana)
- [ ] Фаза 6 — Production hardening (graceful shutdown, rate limiting, health checks, Helm)
