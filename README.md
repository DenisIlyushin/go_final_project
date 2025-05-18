# -go_final_project

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

# Заметки разработки

Создание файла переменных в корне проекта
```bash
cp .env.example .env
```
после отредактировать файл добавив нужные значения

Запуск
```bash
go run .
```
после идём  [http://localhost:{TODO_PORT}/]() или [http://localhost:7540/](http://localhost:7540/),
если переменная TODO_PORT не указана.


```test
.
├── .env                 # настройки окружения (TODO_PORT, TODO_DBFILE, TODO_PASSWORD)
├── go.mod               # модуль проекта
├── main.go              # инициализация и запуск
├── constants
│   └── constants.go     # общие константы
├── database
│   └── database.go      # инициализация SQLite и схема
├── models
│   └── task.go          # структура Task
├── helpers
│   └── nextdate.go      # функция NextDate и утилиты даты
├── handlers
│   ├── add_task.go      # POST /api/task
│   ├── get_task.go      # GET /api/task
│   ├── update_task.go   # PUT /api/task
│   ├── delete_task.go   # DELETE /api/task
│   ├── done_task.go     # POST /api/task/done
│   ├── list_tasks.go    # GET /api/tasks
│   └── nextdate.go      # GET /api/nextdate
├── server
│   └── server.go        # маршрутизация и запуск HTTP-сервера
├── web                  # фронтенд
└── tests                # тесты (оставить без изменений)
```

# Тестирование
## 1
```bash
go test -run ^TestApp$ ./tests
```
## 2
```bash
go test -run ^TestDB$ ./tests
```
## 3
```bash
go test -run ^TestNextDate$ ./tests
```
## 4
```bash
go test -run ^TestAddTask$ ./tests
```
## 5
```bash
go test -run ^TestTasks$ ./tests
```
## 6
```bash
go test -run ^TestEditTask$ ./tests
```
## 7
```bash

```
## 8
```bash

```
