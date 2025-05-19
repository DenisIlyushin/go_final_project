package database

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

const (
	ErrOpenDB        = "opening database"
	ErrConnectDB     = "connecting to database"
	ErrMigrationExec = "migration exec"
)

const (
	createSchedulerTable = `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL DEFAULT '',
        title TEXT NOT NULL DEFAULT '',
        comment TEXT NOT NULL DEFAULT '',
        repeat TEXT NOT NULL DEFAULT ''
    );`

	createSchedulerDateIndex = `
    CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);`
)

// Database представляет обёртку над *sql.DB.
type Database struct {
	db *sql.DB
}

// NewDatabase возвращает новый объект Database, оборачивающий переданный *sql.DB.
func NewDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

// OpenDatabase открывает соединение с SQLite-базой по указанному path,
// проверяет соединение, выполняет миграции (создание таблицы и индекса)
// и возвращает готовую обёртку Database.
// Для драйвера modernc.org/sqlite используем имя драйвера "sqlite".
func OpenDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrOpenDB, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", ErrConnectDB, err)
	}

	if _, err := db.Exec(createSchedulerTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", ErrMigrationExec, err)
	}
	if _, err := db.Exec(createSchedulerDateIndex); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", ErrMigrationExec, err)
	}

	return NewDatabase(db), nil
}

// Close закрывает соединение с базой данных.
func (d *Database) Close() error {
	return d.db.Close()
}

// Exec выполняет SQL-запрос с аргументами args,
// возвращает результат выполнения sql.Result или ошибку.
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}
