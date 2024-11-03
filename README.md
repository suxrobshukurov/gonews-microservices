# Обзор проекта микросервисов для новостей


## Запуск проекта

Проект запускается с помощью скрипта `run.sh`, который компилирует и запускает все необходимые микросервисы:

```bash
bash run.sh
```

### Настройки базы данных

Для корректной работы проекта требуется база данных PostgreSQL, с указанной строкой подключения в файле `.env`. Шаблон `.env-example` содержит примеры необходимых переменных.

## Архитектура микросервисов

### Основной обработчик запросов: APIGateway

APIGateway обрабатывает все входящие HTTP-запросы и направляет их на соответствующие микросервисы. В частности, он предоставляет следующие API-эндпоинты:

- **`GET /news`**: Получить список новостей с пагинацией по умолчанию стоит вывод 10 новостей и первая страница.
- **`GET /news?page=`**: Получить список новостей с пагинацией, используя параметр `page` можно указать нужную страницу.
- **`GET /news/filter?s=`**: Получает список новостей по сопводению к строке title, используя параметр `s`.
- **`GET /news/id?id=`**: Получить детали новости по ID, используя параметр `id`.
- **`POST /news/comment`**: Добавить комментарий к новости. формат тела запроса: 
```json 
{
  "PostID": "id новости", 
  "Content": "текст комментария",
  "ParentID": "id родительского комментария", 
  "AddTime": "время добавления комментария" 
}
```

## Структура проекта

```
./
├── .gitignore
├── APIGateway
│   ├── Dockerfile
│   ├── cmd
│   │   └── server
│   │       ├── apigateway.log
│   │       ├── server.exe
│   │       └── server.go
│   ├── go.mod
│   ├── go.sum
│   └── pkg
│       ├── api
│       │   └── api.go
│       └── models
│           └── models.go
├── Cenzor
│   ├── Dockerfile
│   ├── cmd
│   │   └── server
│   │       ├── cenzor.log
│   │       ├── server.exe
│   │       └── server.go
│   ├── go.mod
│   ├── go.sum
│   └── pkg
│       ├── api
│       │   ├── api.go
│       │   └── api_test.go
│       └── models
│           └── models.go
├── Comments
│   ├── Dockerfile
│   ├── cmd
│   │   └── server
│   │       ├── .env
│   │       ├── comments.log
│   │       ├── server.exe
│   │       └── server.go
│   ├── go.mod
│   ├── go.sum
│   └── pkg
│       ├── api
│       │   ├── api.go
│       │   └── api_test.go
│       ├── db
│       │   ├── db.go
│       │   └── db_test.go
│       └── models
│           └── models.go
├── Gonews
│   ├── Dockerfile
│   ├── cmd
│   │   └── server
│   │       ├── .env
│   │       ├── config.json
│   │       ├── gonews.log
│   │       ├── server.exe
│   │       └── server.go
│   ├── go.mod
│   ├── go.sum
│   └── pkg
│       ├── api
│       │   ├── api.go
│       │   └── api_test.go
│       ├── paginate
│       │   └── paginate.go
│       ├── rss
│       │   ├── rss.go
│       │   └── rss_test.go
│       └── storage
│           ├── memdb
│           │   ├── memdb.go
│           │   └── memdb_test.go
│           └── postgres
│               ├── postgres.go
│               └── postgres_test.go
├── README.md
└── run.sh
```

## Примечания по архитектуре

- Каждый микросервис (APIGateway, Gonews, Comments, Cenzor) выполняет свою уникальную задачу в экосистеме, что позволяет обеспечить модульность и масштабируемость.
