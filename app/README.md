# Multi-Level Cache Service

Multi-Level Cache Service - это экспериментальное приложение на Go, демонстрирующее работу многоуровневого кэша. Сервис позволяет прозрачно использовать несколько слоёв хранения данных и может интегрироваться с клиентами на Spring через REST API.

## Содержание
- [Структура репозитория](#структура-репозитория)
- [Архитектура](#архитектура)
- [Основные операции](#основные-операции)
- [REST API](#rest-api)
- [Конфигурация](#конфигурация)
- [Сборка](#сборка)
- [Запуск тестов](#запуск-тестов)

## Структура репозитория
- `cmd/` — точка входа приложения. В папке `server` находится `main.go`.
- `api/` — DTO и мапперы для обмена данными между слоями.
- `internal/`
  - `cache/` — логика многослойного кэша и провайдеры.
  - `integration/` — получение данных из внешних сервисов при промахах.
  - `manager/` — высокоуровневый менеджер, упрощающий работу с кэшем.
- `configs/` — пример конфигурации `cache.yml`.
- `tests/` — интеграционные тесты (например, для RocksDB клиента).

## Архитектура
Система кэширования построена по принципу каскада с несколькими уровнями хранения. Каждый уровень может использовать собственный провайдер:

- **L0 (Ristretto)** — самый быстрый in-memory кэш, предназначенный для «горячих» данных.
- **L1 (Redis)** — быстрый кэш в оперативной памяти. Используется для хранения часто запрашиваемых значений.
- **L2 (RocksDB)** — локальный дисковый кэш. Позволяет переживать перезапуски сервиса.
- **L3 (API Gateway)** — внешний источник данных, из которого запрашиваются значения при промахах во всех слоях.

Клиент взаимодействует с кэшем как с единой системой. Если значение найдено на одном из уровней, оно возвращается и сохраняется в более высоких слоях (каскадное обновление).

## Основные операции
### Получение значения (GET)
1. Формируется ключ `<cache>:<key>` и выполняется поиск в L0.
2. При промахе запрос идёт в L1, далее в L2.
3. Если значение найдено на более медленном уровне, оно копируется в вышестоящие уровни.
4. При отсутствии значения во всех слоях выполняется запрос к L3 (внешний API).

### Сохранение значения (PUT)
Данные записываются последовательно на все активные уровни кэша. TTL может задаваться индивидуально для каждого уровня.

### Удаление значения (DELETE)
Значение удаляется из всех уровней. При необходимости отправляется запрос во внешний API.

### Пакетные операции (BATCH)
Сервис поддерживает пакетные GET/PUT/DELETE для оптимизации сетевых вызовов. Каждый элемент пакета обрабатывается так же, как одиночный запрос.

## REST API
Ниже приведены основные эндпоинты сервиса.



### Пакетные запросы
```
POST /api/v1/cache/get_all
POST /api/v1/cache/put_all
POST /api/v1/cache/evict_all
```
Формат тела запроса содержит массив объектов с полями `c` (cacheName) и `k` (key). Для PUT также передаётся `v` (value).

Ниже приведены примеры тел запросов и ответов.

#### Get-All

Тело запроса

```json
{
  "requests": [
    {"c": "user", "k": "1"},
    {"c": "user", "k": "2"}
  ]
}
```

Ответ

```json
{
  "results": [
    {"c": "user", "k": "1", "v": {"name": "Ann"}, "f": true},
    {"c": "user", "k": "2", "v": null, "f": false}
  ]
}
```

#### Put-All

Тело запроса

```json
{
  "requests": [
    {"c": "user", "k": "1", "v": {"name": "Ann"}},
    {"c": "user", "k": "2", "v": {"name": "Bob"}}
  ]
}
```

Ответ — HTTP 200 без тела

#### Evict-All

Тело запроса

```json
{
  "requests": [
    {"c": "user", "k": "1"},
    {"c": "user", "k": "2"}
  ]
}
```

Ответ — HTTP 200 без тела

## Конфигурация
Конфигурационный файл `configs/cache.yml` описывает провайдеры, порядок слоёв и параметры отдельных кэшей. Пример фрагмента:

```yaml
providers:
  - name: "ristretto-l0"
    type: "ristretto"
    numCounters: 1_000_000
    bufferItems: 64
    maxCost: 64MiB
    defaultTTL: 15s

  - name: "redis-l1"
    type: "redis"
    host: localhost
    port: 6370
    password: "12345"
    db: 0
    poolSize: 10
    timeout: 5s

  - name: "rocksdb-l2"
    type: "rocksdb"
    path: "/path"
    createIfMissing: true
    maxOpenFiles: 100
    blockSize: 64MiB
    blockCache: 64MiB
    writeBufferSize: 64MiB

layers:
  - name: "ristretto-l0"
    mode: "enabled"
  - name: "redis-l1"
    mode: "enabled"
  - name: "rocksdb-l2"
    mode: "enabled"

caches:
  - name: user
    prefix: "u"
    layers:
      - enabled: true
        ttl: 30s
      - enabled: true
        ttl: 10m
      - enabled: true
        ttl: 6h
    Api:
      enabled: true
      getBatch:
        url: "localhost:8080/user"
        prop: "id"
```

Конфигурация валидируется при запуске приложения (см. пакет `internal/cache/config`).

## Контракт getBatch

Эндпоинт, указанный в конфигурации в разделе `Api.getBatch`, отвечает за
получение данных из внешнего сервиса. Сервис обращается к нему методом
`POST` и передаёт массив ключей в теле запроса. Имя поля определяется
параметром `prop`, тип элементов зависит от `keyType`.

```json
{ "id": ["1", "2"] }
```

Для `keyType: number` вместо строк передаются числа.

Ответ должен представлять собой JSON‑объект, где ключи — это строки с
найденными идентификаторами, а значения содержат соответствующие
объекты. Ключи, для которых запись отсутствует, в ответе не
возвращаются.

```json
{
  "1": {"name": "Ann"},
  "2": {"name": "Bob"}
}
```

Пример с отсутствующим значением:

```json
{
  "1": {"name": "Ann"}
}
```

Код ответа должен быть `200` (или другой `2xx`). Любой иной код
считается ошибкой, и соответствующая группа ключей будет помечена как
`Skipped`.

## Сборка
Для сборки требуется Go 1.24+. Пример последовательности действий:

```bash
go mod download

go build ./cmd/server
```

## Запуск тестов

```bash
go test ./...
```

## Метрики

Сервис экспортирует метрики в формате Prometheus. После запуска приложения они
доступны по HTTP на пути `/metrics` (по умолчанию порт `9080`). Достаточно
перейти в браузере по ссылке:

```
http://localhost:9080/metrics
```

Проверка health
```
http://localhost:9080/metrics/health
{"status":"UP"}
```

Для сбора статистики можно настроить Prometheus, добавив в конфигурацию
следующий scrape target:

```yaml
scrape_configs:
  - job_name: cache-service
    static_configs:
      - targets: ['localhost:8080']

