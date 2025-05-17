package database

import (
	"fmt"
	"github.com/DenisIlyushin/go_final_project/models"
)

const (
	ErrAddTaskPrepare = "не удалось подготовить запрос добавления задачи"
	ErrAddTaskExec    = "не удалось выполнить запрос добавления задачи"
	ErrAddTaskLastID  = "не удалось получить ID добавленной задачи"
)

// AddTask сохраняет новую задачу в таблицу scheduler.
// Принимает структуру models.Task и возвращает ID добавленной записи или ошибку.
func (d *Database) CreateTask(task models.Task) (int64, error) {
	stmt, err := d.db.Prepare("INSERT INTO scheduler(date, title, comment, repeat) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", ErrAddTaskPrepare, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", ErrAddTaskExec, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", ErrAddTaskLastID, err)
	}
	return id, nil
}
