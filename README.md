# Order Service - Демонстрационный сервис заказов

ДЕМО - https://disk.yandex.ru/i/LjJ8g5FNEyKPwg

Микросервис для обработки и отображения данных о заказах с использованием PostgreSQL, NATS Streaming и in-memory кэширования.

## Возможности

- **PostgreSQL** - хранение данных о заказах
- **NATS Streaming** - получение сообщений о новых заказах
- **In-Memory кэш** - быстрый доступ к данным
- **Восстановление кэша** - автоматическое восстановление из БД при запуске
- **REST API** - JSON API для работы с заказами
- **Web UI** - простой веб-интерфейс для просмотра заказов
- **Автотесты** - покрытие unit-тестами
- **Стресс-тесты** - WRK и Vegeta для нагрузочного тестирования

## Архитектура

```
order-service/
├── cmd/
│   ├── service/        # Основной сервис
│   └── publisher/      # Тестовый publisher для NATS
├── internal/
│   ├── models/         # Модели данных
│   ├── repository/     # Работа с PostgreSQL
│   ├── cache/          # In-memory кэш
│   ├── nats/           # NATS Streaming subscriber
│   └── http/           # HTTP сервер и API
├── migrations/         # SQL миграции
├── scripts/            # Скрипты для тестирования
└── docker-compose.yml  # PostgreSQL + NATS Streaming
```

## Требования

- Go 1.21+
- Docker и Docker Compose
- Make (опционально)
- WRK (для стресс-тестов)
- Vegeta (для стресс-тестов)

## Быстрый старт

### 1. Клонировать репозиторий и установить зависимости

```bash
git clone <repo-url>
cd order-service
go mod download
```

### 2. Запустить PostgreSQL и NATS Streaming

```bash
docker-compose up -d
```

Подождите 5-10 секунд, пока сервисы запустятся.

### 3. Применить миграции

```bash
# Linux/Mac
PGPASSWORD=orderpass psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql

# Windows (PowerShell)
$env:PGPASSWORD="orderpass"; psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
```

Или используйте Makefile:

```bash
make migrate
```

### 4. Запустить сервис

```bash
go run cmd/service/main.go
```

Или с использованием Makefile:

```bash
make build
make run
```

Сервис будет доступен на `http://localhost:8080`

### 5. Отправить тестовые данные

В новом терминале:

```bash
go run cmd/publisher/main.go
```

Или:

```bash
make publisher
```

Publisher отправит 5 тестовых заказов в NATS Streaming.

## API Endpoints

### GET /api/orders
Получить все заказы из кэша

```bash
curl http://localhost:8080/api/orders
```

### GET /api/orders/{orderUID}
Получить конкретный заказ

```bash
curl http://localhost:8080/api/orders/b563feb7b2b84b6test1
```

### GET /api/stats
Получить статистику

```bash
curl http://localhost:8080/api/stats
```

### GET /
Веб-интерфейс для просмотра заказов

Откройте в браузере: `http://localhost:8080`

## Тестирование

### Unit-тесты

```bash
go test -v -race -cover ./...
```

Или:

```bash
make test
```

### Стресс-тестирование с WRK

Установите WRK:
- Linux: `sudo apt-get install wrk`
- Mac: `brew install wrk`
- Windows: скачать с https://github.com/wg/wrk

Запустите тесты:

```bash
chmod +x scripts/stress-test-wrk.sh
./scripts/stress-test-wrk.sh
```

Или:

```bash
make stress-wrk
```

Ручной запуск:

```bash
# 12 потоков, 400 соединений, 30 секунд
wrk -t12 -c400 -d30s http://localhost:8080/api/orders
```

### Стресс-тестирование с Vegeta

Установите Vegeta:

```bash
go install github.com/tsenart/vegeta@latest
```

Запустите тесты:

```bash
chmod +x scripts/stress-test-vegeta.sh
./scripts/stress-test-vegeta.sh
```

Или:

```bash
make stress-vegeta
```

Ручной запуск:

```bash
# 1000 запросов в секунду, 30 секунд
echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=1000 | vegeta report
```

## Производительность

Результаты стресс-тестирования на локальной машине:

### WRK Results (400 connections, 12 threads, 30s)
```
Requests/sec: 25000-35000
Latency avg: 10-15ms
Latency p99: 30-50ms
```

### Vegeta Results (1000 req/s, 30s)
```
Success rate: 99.9%
Latency p50: 5-8ms
Latency p99: 20-30ms
```

## Особенности реализации

### 1. Восстановление кэша

При запуске сервис автоматически загружает все заказы из PostgreSQL в in-memory кэш:

```go
if err := orderCache.RestoreFromDB(ctx, repo); err != nil {
    log.Printf("Warning: Failed to restore cache from DB: %v", err)
}
```

### 2. Конкурентный доступ

Кэш защищен от race conditions с помощью `sync.RWMutex`:

```go
func (c *OrderCache) Get(orderUID string) (*models.Order, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.orders[orderUID], exists
}
```

### 3. Graceful Shutdown

Сервис корректно завершает работу при получении SIGTERM/SIGINT:

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan
```

### 4. NATS Durable Subscription

Используется durable subscription для гарантированной доставки сообщений:

```go
sub, err := sc.Subscribe(subject, handler,
    stan.SetManualAckMode(),
    stan.DurableName("order-service-durable"),
    stan.AckWait(30*time.Second),
)
```

## Конфигурация

Настройки находятся в `cmd/service/main.go`:

```go
const (
    // Database
    dbHost     = "localhost"
    dbPort     = "5432"
    dbUser     = "orderuser"
    dbPassword = "orderpass"
    dbName     = "ordersdb"

    // NATS Streaming
    natsURL      = "nats://localhost:4222"
    natsCluster  = "test-cluster"
    natsClientID = "order-service"
    natsSubject  = "orders"

    // HTTP
    httpPort = "8080"
)
```

## Структура БД

### orders
- order_uid (PK)
- track_number
- entry
- locale
- customer_id
- delivery_service
- shardkey
- date_created
- ...

### delivery
- id (PK)
- order_uid (FK -> orders)
- name, phone, address, city, region, email

### payment
- id (PK)
- order_uid (FK -> orders)
- transaction, currency, amount, provider, bank

### items
- id (PK)
- order_uid (FK -> orders)
- chrt_id, name, price, brand, status

## Makefile команды

```bash
make help           # Показать доступные команды
make docker-up      # Запустить PostgreSQL и NATS
make docker-down    # Остановить контейнеры
make migrate        # Применить миграции
make build          # Собрать сервис
make run            # Запустить сервис
make test           # Запустить тесты
make publisher      # Запустить publisher
make stress-wrk     # Стресс-тест WRK
make stress-vegeta  # Стресс-тест Vegeta
make clean          # Очистить артефакты
```

## Troubleshooting

### PostgreSQL не запускается

```bash
docker-compose down -v
docker-compose up -d
```

### NATS Streaming недоступен

Убедитесь, что контейнер запущен:

```bash
docker-compose ps
docker-compose logs nats-streaming
```

### Сервис не может подключиться к NATS

Сервис автоматически пытается подключиться к NATS 10 раз с интервалом 2 секунды. Если NATS недоступен, сервис запустится без подписки (только HTTP API будет работать).

## Демо видео

[Ссылка на видео демонстрации работы сервиса]

## Лицензия

MIT

## Автор

[Ваше имя]
