# go_final_project

Это веб-сервер «Планировщик задач» (аналог TODO-листа) на Go с поддержкой:

- CRUD-API для задач (добавление, просмотр, редактирование, удаление, отметка выполненным)
- Фронтенда (HTML/CSS/JS) в папке `web`
- Создания повторяющихся задач по правилам:
    - каждый N-й день (`d N`)
    - ежегодно (`y`)
    - по дню(дням) недели (`w …`)*
    - по дню(дням) месяца и месяцам (`m …`)*
- Поиска задач по заголовку и комментарию*
- Аутентификации через простой JWT-токен*
- Сборки приложения через docker compose
---

## Список выполненных «звёздочных» заданий

1. **Шаг 1★** — определение порта через переменную окружения `TODO_PORT`
2. **Шаг 2★** — определение пути к файлу SQLite через `TODO_DBFILE`
3. **Шаг 3★** — поддержка правил повторения `w` (дни недели) и `m` (дни/месяцы)
4. **Шаг 5★** — поиск задач по `search` в заголовке/комментарии и по дате
5. **Шаг 8★** — базовая аутентификация (JWT-кука, эндпоинт `/api/signin`)
6. **Шаг 8★** — поддержка docker

# Структура проекта
```test
.
├── .dockerignore         # Файл для исключения при сборке Docker-образа
├── .env.example          # Пример файла с переменными окружения
├── Dockerfile            # Инструкция по сборке Docker-образа
├── README.md             # Документация и инструкции по запуску
├── auth
│   └── auth.go           # Логика JWT-аутентификации и middleware
├── config
│   ├── config.go         # Загрузка настроек из `.env`
│   └── constants.go      # Общие константы (формат даты, переменные окружения)
├── database
│   ├── database.go       # Инициализация SQLite, схема (таблица, индекс)
│   └── task_crud.go      # CRUD-функции работы с задачами
├── docker-compose.yml    # Пример запуска через Docker Compose
├── go.mod                # Go-модуль и зависимости
├── go.sum                # Контрольные суммы зависимостей
├── handlers
│   └── task.go           # HTTP-обработчики для модели `task`
├── main.go               # Точка входа: инициализация и запуск сервера
├── models
│   └── task.go           # Структура Task для JSON и БД
├── scheduler.db          # Файл SQLite (генерируется при первом запуске)
├── server
│   ├── router.go         # Настройка маршрутов и middleware
│   └── server.go         # Запуск HTTP-сервера на порту из конфига
├── tests                 # Тесты от учебного проекта
├── utils
│   ├── nextdate.go       # Алгоритмы для расчёта следующей даты
│   └── validators.go     # Валидация форматов даты/входных данных
└── web                   # Клиентская часть (HTML/CSS/JS)
```

# Локальный запуск приложения

Создание файла переменных в корне проекта
```bash
cp .env.example .env
```
после отредактировать файл добавив нужные значения

Запуск
```bash
go mod tidy
go run .
```
после идём  [http://localhost:{TODO_PORT}/]() или [http://localhost:7540/](http://localhost:7540/),
если переменная TODO_PORT не указана.

# Запуск через docker compose


Создание файла переменных в корне проекта
```bash
cp .env.example .env
```
после отредактировать файл добавив нужные значения

Запуск

```bash
docker compose up -d
```
после идём  [http://localhost:{TODO_PORT}/]() или [http://localhost:7540/](http://localhost:7540/),
если переменная TODO_PORT не указана.

# Запуск тестов
В проекте есть набор тестов в папке tests.

Перед запуском откройте tests/settings.go и при необходимости скорректируйте там:
```go
var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = `XXX`               // значение токена, которое сервер возвратил из `/api/signin` и которое хранится в куке `token`
```

запуск тестов
```go
go test ./tests
```

Запуск отдельных тестов
#### Шаг 1
```bash
go test -run ^TestApp$ ./tests
```
#### Шаг 2
```bash
go test -run ^TestDB$ ./tests
```
#### Шаг 3
```bash
go test -run ^TestNextDate$ ./tests
```
#### Шаг 4
```bash
go test -run ^TestAddTask$ ./tests
```
#### Шаг 5
```bash
go test -run ^TestTasks$ ./tests
```
#### Шаг 6
```bash
go test -run ^TestEditTask$ ./tests
```
#### Шаг 7
```bash
go test -run ^TestDone$ ./tests
go test -run ^TestDelTask$ ./tests
```