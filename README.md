# Dynamic Customer Segmentation

Микросервис для работы с сегментами пользователей

## Используемые технологии

- net/http (в качестве веб-сервера)
- chi (в качестве HTTP router)
- PostgreSQL (в качестве БД)
- golang-migrate/migrate (для миграций БД)
- zerolog (для логгирования)
- Docker (для запуска сервиса)

## Usage

Запустить сервис можно командой:

```bash
 $ docker compose up --build
```

## Примеры запросов

### Создание сегмента

```bash
curl --request POST --url 'http://localhost:80/api/v1/segment/create' \
--header "Content-Type: application/json" \
--data '{
    "slug": "AVITO_TEST_SEGMENT"
}'
```

Ответ:

```json
{
    "status": "OK"
}
```

### Удаление сегмента

```bash
curl --request POST --url 'http://localhost:80/api/v1/segment/delete' \
--header "Content-Type: application/json" \
--data '{
    "slug": "AVITO_TEST_SEGMENT"
}'
```

Ответ:

```json
{
    "status": "OK"
}
```

### Создание сегмента и добавление его проценту пользователей

```bash
curl --location 'http://localhost:80/api/v1/segment/create/enroll' \
--header 'Content-Type: application/json' \
--data '{"slug": "AVITO_RANDOM_SEGMENT", "percent": 5}'
```

Ответ:

```json
{
    "user_ids": [
        1219,
        1299,
        1105,
        1235,
        1147,
        1154,
        1024,
        1042,
        1165,
        1216,
        1102,
        1217,
        1296,
        1253,
        1055
    ]
}
```

### Получение всех активных сегментов

```bash
curl --location 'http://localhost:80/api/v1/segments/active'
```

Ответ:

```json
{
    "segments": [
        {
            "slug": "AVITO_VOICE_CHAT",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_SALE_30",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_SALE_50",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_PERFORMANCE_VAS",
            "created_at": "2023-08-28T18:02:39.642461Z"
        }
    ]
}
```

### Получение всех сегментов (в том числе удалённых)

```bash
curl --location 'http://localhost:80/api/v1/segments'
```

Ответ:

```json
{
    "segments": [
        {
            "slug": "AVITO_VOICE_CHAT",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_SALE_30",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_SALE_50",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_PERFORMANCE_VAS",
            "created_at": "2023-08-28T18:02:39.642461Z"
        },
        {
            "slug": "AVITO_ABC",
            "created_at": "2023-08-28T18:03:08.537641Z"
        },
        {
            "slug": "AVITO_TEST_SEGMENT",
            "created_at": "2023-08-28T18:28:00.250362Z",
            "deleted_at": "2023-08-28T18:31:27.573511Z"
        }
    ]
}
```

### Добавление и удаление пользователя из сегментов

```bash
curl --request POST --location 'http://localhost:80/api/v1/user/update' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 1012,
    "add_segments": [
        {"slug": "AVITO_EXAMPLE"},
        {"slug": "AVITO_SEGMENT_WITH_EXPIRATION", "expires_at": "2023-08-28T21:39:28.168792Z"}
    ],
    "remove_segments": [
        {"slug": "AVITO_TO_BE_REMOVED"}
    ]
}'
```

Ответ:

```json
{
    "status": "OK"
}
```

### Получение активных сегментов пользователя

```bash
curl --location --request GET 'http://localhost:80/api/v1/user/segments' \
--header 'Content-Type: application/json' \
--data '{"user_id": 1012}'
```

Ответ:

```json
{
    "segments": [
        {
            "slug": "AVITO_ABC",
            "added_at": "2023-08-28T18:03:08.548867Z"
        },
        {
            "slug": "AVITO_EXAMPLE",
            "added_at": "2023-08-28T18:37:55.457516Z"
        },
        {
            "slug": "AVITO_SEGMENT_WITH_EXPIRATION",
            "added_at": "2023-08-28T18:37:55.457516Z",
            "expires_at": "2023-08-28T21:39:28.168792Z"
        }
    ]
}
```

### Получение отчёта в CSV

```bash
curl --location --request GET 'http://localhost:80/api/v1/user/csv' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 1012,
    "from": {
        "month": 1,
        "year": 2023
    },
    "to": {
        "month": 1,
        "year": 2024
    }
}'
```

Ответ:

```json
{
    "link": "http://localhost:80/csv/1012--1.2023-1.2024.csv"
}
```
