# USDT Rate Service

gRPC сервис, который получает курс USDT с биржи Binance, считает итоговые ask/bid
и сохраняет каждый полученный курс в PostgreSQL.

## Возможности

- gRPC метод `GetRates` — курс ask и bid с меткой времени. Каждый вызов сохраняется в БД.
- gRPC метод `HealthCheck` — проверка работоспособности (доступна ли БД).
- Два способа расчёта курса из стакана (order book):
  - `topN` — цена из позиции N;
  - `avgNM` — среднее значение цен с позиции N по позицию M.
- Graceful shutdown по `SIGINT` / `SIGTERM`.
- Метрики Prometheus, трассировка OpenTelemetry, логи через zap.
- Миграции БД применяются автоматически при старте.

Данные берутся из API Binance: `GET /api/v3/depth?symbol=USDTTRY`.
Запрос выполняется через библиотеку [resty](https://github.com/go-resty/resty).

> Важно: в ответе Binance нет времени, поэтому в качестве метки времени курса
> используется время получения ответа. Пара по умолчанию `USDTTRY`, потому что у
> `USDTRUB` на Binance сейчас пустой стакан. Пару можно поменять через `BINANCE_SYMBOL`.

## Требования

- Go 1.25 или выше
- Docker (или Podman) с docker-compose
- golangci-lint (только для `make lint`)

## Запуск

```bash
git clone <repo-url>
cd gotask
make build
docker-compose up -d --build
```

`docker-compose up` поднимает PostgreSQL и приложение. Приложение стартует само,
ждёт готовности БД и применяет миграции.

Остановить всё: `docker-compose down`.

### Запуск через `docker-compose run`

Команда из задания `docker-compose run --rm app ./app` запускает второй экземпляр
приложения, поэтому сначала нужно остановить сервис `app`, иначе порты будут заняты:

```bash
docker-compose up -d postgres
docker-compose run --rm --service-ports app ./app
```

### Podman

Вместо docker-compose можно использовать `podman-compose`, файлы те же:

```bash
podman-compose up -d --build
```

В `Dockerfile` и `docker-compose.yml` у образов указаны полные имена
(`docker.io/library/...`), чтобы Podman не спрашивал, из какого реестра качать.

### Локальный запуск

Поднимите только PostgreSQL и запустите приложение:

```bash
docker-compose up -d postgres
make run
```

## Проверка работы

В репозитории есть небольшой CLI-клиент `bin/client` (собирается через `make build`):

```bash
# курс методом topN (позиция 1)
./bin/client -method=topN -n=1

# курс методом avgNM (среднее по позициям 1..3)
./bin/client -method=avgNM -n=1 -m=3

# healthcheck
./bin/client -health
```

Или через `grpcurl` (reflection включён):

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"method": "CALCULATION_METHOD_TOP_N", "n": 1}' \
  localhost:50051 rate.v1.RateService/GetRates
```

Проверить, что курс сохранился в БД:

```bash
docker-compose exec postgres psql -U postgres -d ratedb -c "SELECT * FROM rates"
```

Метрики Prometheus: `http://localhost:9090/metrics`.

## Конфигурация

Параметры задаются переменными окружения или флагами запуска.
Если задано и то и другое, флаг важнее.

| Параметр | Переменная | Флаг | По умолчанию |
|---|---|---|---|
| Порт gRPC | `GRPC_PORT` | `-grpc-port` | `50051` |
| Порт метрик | `METRICS_PORT` | `-metrics-port` | `9090` |
| Уровень логов (debug, info, warn, error) | `LOG_LEVEL` | `-log-level` | `info` |
| Автоматические миграции | `AUTO_MIGRATE` | `-auto-migrate` | `true` |
| Хост БД | `DB_HOST` | `-db-host` | `localhost` |
| Порт БД | `DB_PORT` | `-db-port` | `5432` |
| Пользователь БД | `DB_USER` | `-db-user` | `postgres` |
| Пароль БД | `DB_PASSWORD` | `-db-password` | `postgres` |
| Имя БД | `DB_NAME` | `-db-name` | `ratedb` |
| SSL режим БД | `DB_SSLMODE` | `-db-sslmode` | `disable` |
| URL API биржи | `BINANCE_API_URL` | `-exchange-url` | `https://api.binance.com` |
| Торговая пара | `BINANCE_SYMBOL` | `-exchange-symbol` | `USDTTRY` |
| Таймаут запроса к бирже | `BINANCE_TIMEOUT` | `-exchange-timeout` | `5s` |

Пример:

```bash
./bin/app -db-host=localhost -db-name=ratedb -grpc-port=50051
```

## Makefile

| Команда | Что делает |
|---|---|
| `make build` | собирает `bin/app` и `bin/client` |
| `make test` | запускает unit-тесты |
| `make docker-build` | собирает Docker-образ `app:latest` |
| `make run` | запускает приложение локально |
| `make lint` | запускает golangci-lint |
| `make proto` | генерирует gRPC код из `proto/` (нужен `protoc`) |
| `make clean` | удаляет `bin/` |

## Структура проекта

```
cmd/app              точка входа сервиса
cmd/client           CLI-клиент для проверки
internal/config      конфигурация (флаги + env)
internal/client      клиент биржи Binance
internal/calculator  расчёт курса (topN, avgNM)
internal/service     бизнес-логика
internal/grpc        gRPC сервер и обработчики
internal/repository  работа с PostgreSQL
internal/telemetry   метрики и трассировка
migrations           SQL миграции
proto, gen           описание API и сгенерированный код
```
