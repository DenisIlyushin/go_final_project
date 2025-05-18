package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/models"
	"strconv"
	"time"
)

const (
	ErrAddTaskPrepare    = "не удалось подготовить запрос на добавление задачи"
	ErrAddTaskExec       = "не удалось выполнить запрос на добавление задачи"
	ErrAddTaskLastID     = "не удалось получить ID добавленной задачи"
	ErrGetTasksQuery     = "не удалось выполнить запрос на получение задач"
	ErrGetTasksScan      = "не удалось прочитать задачу из результата"
	ErrGetTasksIteration = "ошибка при обходе результатов запроса задач"
	ErrGetTaskParseID    = "не удалось преобразовать ID задачи"
	ErrGetTaskNotFound   = "задача с указанным ID не найдена"
	ErrGetTaskScan       = "не удалось прочитать задачу"
)

// CreateTask сохраняет новую задачу в таблицу scheduler и возвращает сгенерированный ID.
// При возникновении ошибки возвращает один из ErrAddTask*.
func (d *Database) CreateTask(task models.Task) (int64, error) {
	stmt, err := d.db.Prepare(
		"INSERT INTO scheduler(date, title, comment, repeat) VALUES(?, ?, ?, ?)")
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

// GetTasks возвращает список задач из таблицы scheduler, упорядоченных по дате.
// Ограничение числа задач задаётся через config.TasksLimit.
// GetTasks возвращает список задач из таблицы scheduler.
// Если searchQuery пустой, возвращаются все задачи, упорядоченные по дате, с ограничением config.TasksLimit.
// Если searchQuery соответствует формату даты config.DateFormat, выполняется фильтрация по дате.
// Иначе выполняется поиск по вхождению searchQuery в поля title или comment.
// В случае ошибки возвращает одну из ErrGetTasks*.
func (d *Database) GetTasks(searchQuery string) ([]models.Task, error) {
	tasks := make([]models.Task, 0, config.TasksLimit)

	var (
		rows *sql.Rows
		err  error
	)

	switch {
	case searchQuery == "":
		rows, err = d.db.Query(
			"SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?",
			config.TasksLimit)
	default:
		if t, perr := time.Parse(config.DateFormat, searchQuery); perr == nil {
			// поиск по дате
			rows, err = d.db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?",
				t.Format(config.DateFormat),
				config.TasksLimit)
		} else {
			pattern := "%" + searchQuery + "%"
			rows, err = d.db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?",
				pattern,
				pattern,
				config.TasksLimit)
		}
	}
	if err != nil {
		return tasks, fmt.Errorf("%s: %w", ErrGetTasksQuery, err)
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return tasks, fmt.Errorf("%s: %w", ErrGetTasksScan, err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return tasks, fmt.Errorf("%s: %w", ErrGetTasksIteration, err)
	}
	return tasks, nil
}

// GetTask возвращает задачу по её строковому ID.
func (d *Database) GetTask(idStr string) (models.Task, error) {
	var t models.Task
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return t, fmt.Errorf("%s: %w", ErrGetTaskParseID, err)
	}
	row := d.db.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return t, fmt.Errorf("%s: %s", ErrGetTaskNotFound, idStr)
		}
		return t, fmt.Errorf("%s: %w", ErrGetTaskScan, err)
	}

	return t, nil
}

func (d *Database) EditTask(task models.Task) error {

	result, err := d.db.Exec("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task with id %v not found", task.ID)
	}
	return nil
}

// DeleteTask удаляет задачу по её строковому ID.
// Если задача не найдена, возвращает ошибку.
func (d *Database) DeleteTask(idStr string) error {
	// Пытаемся удалить строку
	result, err := d.db.Exec(
		"DELETE FROM scheduler WHERE id = ?", idStr,
	)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу: %w", err)
	}
	// Проверяем, сколько строк затронуто
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось узнать количество удалённых задач: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("задача с ID %s не найдена", idStr)
	}
	return nil
}
