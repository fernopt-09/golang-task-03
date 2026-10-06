# USDT Rate Service

Небольшой сервис, который спрашивает у Binance актуальный курс USDT, считает по стакану цены покупки и продажи и запоминает каждый ответ в PostgreSQL. Курс можно получить по gRPC.

## Что нужно

- Go 1.25+
- Docker (или Podman) с docker-compose
- golangci-lint, если захотите запускать `make lint`

## Как запустить

```bash
git clone <repo-url>
cd gotask
make build
docker-compose up -d --build
```

Поднимутся PostgreSQL и само приложение: оно дождётся базы и сразу применит миграции.
Остановить всё: `docker-compose down`.

### Через `docker-compose run`

Эта команда запускает ещё один экземпляр приложения, поэтому сервис `app` должен быть остановлен, иначе порты окажутся заняты:

```bash
docker-compose up -d postgres
docker-compose run --rm --service-ports app ./app
```

### Podman

Всё то же самое, только команда другая:

```bash
podman-compose up -d --build
```

### Без контейнера для приложения

Поднимите только базу и запустите сервис локально:

```bash
docker-compose up -d postgres
make run
```
