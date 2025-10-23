# Инструкция по установке и запуску Order Service

## Предварительные требования

### 1. Установите Go (версия 1.21 или выше)

**Windows:**
- Скачайте с https://golang.org/dl/
- Установите и добавьте в PATH

**Linux:**
```bash
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Mac:**
```bash
brew install go
```

### 2. Установите Docker и Docker Compose

**Windows/Mac:**
- Docker Desktop: https://www.docker.com/products/docker-desktop

**Linux:**
```bash
sudo apt-get update
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER
```

### 3. Установите PostgreSQL клиент (psql)

**Windows:**
- Скачайте PostgreSQL с https://www.postgresql.org/download/windows/
- Или используйте Chocolatey: `choco install postgresql`

**Linux:**
```bash
sudo apt-get install postgresql-client
```

**Mac:**
```bash
brew install postgresql
```

### 4. (Опционально) Установите инструменты для стресс-тестов

**WRK:**

Linux:
```bash
sudo apt-get install wrk
```

Mac:
```bash
brew install wrk
```

Windows:
- Скачайте с https://github.com/wg/wrk

**Vegeta:**

```bash
go install github.com/tsenart/vegeta@latest
```

## Пошаговая установка

### Шаг 1: Клонируйте репозиторий

```bash
git clone <your-repo-url>
cd order-service
```

### Шаг 2: Установите зависимости Go

```bash
go mod download
```

### Шаг 3: Запустите PostgreSQL и NATS Streaming

```bash
docker-compose up -d
```

Проверьте, что контейнеры запущены:

```bash
docker-compose ps
```

Вы должны увидеть:
- `orders_postgres` - running
- `orders_nats` - running

### Шаг 4: Дождитесь готовности PostgreSQL

```bash
# Подождите 5-10 секунд
sleep 10

# Проверьте подключение
docker-compose exec postgres pg_isready -U orderuser
```

### Шаг 5: Примените миграции базы данных

**Linux/Mac:**
```bash
PGPASSWORD=orderpass psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
```

**Windows (PowerShell):**
```powershell
$env:PGPASSWORD="orderpass"
psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
```

**Windows (CMD):**
```cmd
set PGPASSWORD=orderpass
psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
```

Или используйте Docker:
```bash
docker-compose exec postgres psql -U orderuser -d ordersdb -f /migrations/001_init_schema.sql
```

### Шаг 6: Соберите проект

```bash
go build -o bin/service cmd/service/main.go
go build -o bin/publisher cmd/publisher/main.go
```

**Windows:**
```bash
go build -o bin/service.exe cmd/service/main.go
go build -o bin/publisher.exe cmd/publisher/main.go
```

### Шаг 7: Запустите сервис

**Linux/Mac:**
```bash
./bin/service
```

**Windows:**
```bash
bin\service.exe
```

Или напрямую:
```bash
go run cmd/service/main.go
```

Вы должны увидеть:
```
Starting Order Service...
Connecting to PostgreSQL...
Successfully connected to PostgreSQL
Restoring cache from database...
Cache restored successfully. Total orders in cache: 0
Successfully connected to NATS Streaming
Service started successfully!
HTTP server: http://localhost:8080
Cache size: 0 orders
```

### Шаг 8: Откройте веб-интерфейс

Откройте браузер и перейдите на: http://localhost:8080

Вы должны увидеть главную страницу с:
- Статистикой кэша (0 заказов)
- Поле поиска заказов
- Кнопку "Load All Orders"

### Шаг 9: Отправьте тестовые данные

В новом терминале:

**Linux/Mac:**
```bash
./bin/publisher
```

**Windows:**
```bash
bin\publisher.exe
```

Или:
```bash
go run cmd/publisher/main.go
```

Вы должны увидеть:
```
Starting NATS Publisher...
Connected to NATS Streaming
Published order: b563feb7b2b84b6test1
Published order: b563feb7b2b84b6test2
Published order: b563feb7b2b84b6test3
Published order: b563feb7b2b84b6test4
Published order: b563feb7b2b84b6test5
All orders published successfully
```

В логах сервиса:
```
Received message: {...}
Order b563feb7b2b84b6test1 processed successfully
...
```

### Шаг 10: Проверьте работу

1. **Обновите веб-страницу** - статистика должна показать 5 заказов

2. **Нажмите "Load All Orders"** - должны отобразиться все 5 заказов

3. **Введите ID заказа** `b563feb7b2b84b6test1` и нажмите "Search Order"

4. **Проверьте API:**

```bash
# Все заказы
curl http://localhost:8080/api/orders

# Конкретный заказ
curl http://localhost:8080/api/orders/b563feb7b2b84b6test1

# Статистика
curl http://localhost:8080/api/stats
```

## Тестирование восстановления кэша

### Шаг 1: Остановите сервис
Нажмите `Ctrl+C` в терминале с сервисом

### Шаг 2: Запустите сервис заново

```bash
./bin/service
```

Вы должны увидеть:
```
Restoring cache from database...
Cache restored successfully. Total orders in cache: 5
```

Это подтверждает, что кэш успешно восстановился из БД!

## Запуск тестов

```bash
# Unit-тесты
go test -v ./...

# С покрытием
go test -v -cover ./...

# Race detector
go test -v -race ./...
```

## Стресс-тестирование

### WRK

```bash
# Сделайте скрипт исполняемым (Linux/Mac)
chmod +x scripts/stress-test-wrk.sh

# Запустите
./scripts/stress-test-wrk.sh
```

Или вручную:
```bash
wrk -t12 -c400 -d30s http://localhost:8080/api/orders
```

### Vegeta

```bash
# Сделайте скрипт исполняемым (Linux/Mac)
chmod +x scripts/stress-test-vegeta.sh

# Запустите
./scripts/stress-test-vegeta.sh
```

Или вручную:
```bash
echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=1000 | vegeta report
```

## Остановка и очистка

### Остановить сервис
Нажмите `Ctrl+C`

### Остановить Docker контейнеры
```bash
docker-compose down
```

### Полная очистка (включая данные)
```bash
docker-compose down -v
rm -rf bin/
```

## Troubleshooting

### Ошибка: "Failed to connect to database"

**Проблема:** PostgreSQL не готов или не запущен

**Решение:**
```bash
# Проверьте статус
docker-compose ps

# Перезапустите
docker-compose restart postgres

# Проверьте логи
docker-compose logs postgres
```

### Ошибка: "Failed to connect to NATS Streaming"

**Проблема:** NATS не запущен или не готов

**Решение:**
```bash
# Проверьте статус
docker-compose ps

# Перезапустите
docker-compose restart nats-streaming

# Проверьте логи
docker-compose logs nats-streaming
```

Сервис автоматически переподключится к NATS при восстановлении связи.

### Ошибка: "dial tcp [::1]:5432: connect: connection refused"

**Проблема:** PostgreSQL слушает только IPv4, а приложение пытается подключиться через IPv6

**Решение:** Измените `localhost` на `127.0.0.1` в `cmd/service/main.go`

### Порт 8080 уже занят

**Решение:** Измените порт в `cmd/service/main.go`:
```go
const httpPort = "8081"
```

### Миграции не применяются

**Решение:** Используйте Docker:
```bash
docker cp migrations/001_init_schema.sql orders_postgres:/tmp/
docker-compose exec postgres psql -U orderuser -d ordersdb -f /tmp/001_init_schema.sql
```

## Полезные команды

### Подключиться к PostgreSQL

```bash
docker-compose exec postgres psql -U orderuser -d ordersdb
```

### Посмотреть данные в БД

```sql
-- Все заказы
SELECT order_uid, track_number, customer_id FROM orders;

-- Количество заказов
SELECT COUNT(*) FROM orders;

-- Полная информация о заказе
SELECT * FROM orders WHERE order_uid = 'b563feb7b2b84b6test1';
```

### Очистить данные

```sql
TRUNCATE orders CASCADE;
```

### Посмотреть логи NATS

```bash
docker-compose logs -f nats-streaming
```

### Перезапустить сервисы

```bash
docker-compose restart
```

## Готово!

Теперь у вас полностью рабочий сервис заказов с:
- ✅ PostgreSQL базой данных
- ✅ NATS Streaming подпиской
- ✅ In-memory кэшем
- ✅ REST API
- ✅ Веб-интерфейсом
- ✅ Автотестами
- ✅ Стресс-тестами

Для демонстрации запишите короткое видео, показывающее:
1. Запуск сервиса
2. Отправку данных через publisher
3. Просмотр заказов в веб-интерфейсе
4. Поиск конкретного заказа
5. Перезапуск сервиса и восстановление кэша
6. Результаты стресс-тестов
