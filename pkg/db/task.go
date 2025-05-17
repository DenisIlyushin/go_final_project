package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func Tasks(limit int, search string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		if t, errDate := time.Parse("02.01.2006", search); errDate == nil {
			date := t.Format("20060102")
			rows, err = DB.Query(`
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE date = ?
				ORDER BY date
				LIMIT ?`, date, limit)
		} else {
			words := strings.Fields(search)
			if len(words) == 0 {
				return defaultTasks(limit)
			}

			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE `
			args := []any{}
			for i, word := range words {
				if i > 0 {
					query += " AND "
				}
				query += `(title LIKE ? COLLATE NOCASE OR comment LIKE ? COLLATE NOCASE)`
				p := "%" + word + "%"
				args = append(args, p, p)
			}
			query += " ORDER BY date LIMIT ?"
			args = append(args, limit)

			rows, err = DB.Query(query, args...)
		}
	} else {
		return defaultTasks(limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}

	return list, nil
}

func AddTask(task *Task) (int64, error) {
	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetTask(id string) (*Task, error) {
	row := DB.QueryRow(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?`, id)

	var t Task
	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Задача не найдена")
		}
		return nil, err
	}
	return &t, nil
}

func UpdateDate(date, id string) error {
	res, err := DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", date, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func DeleteTask(id string) error {
	_, err := DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}

func UpdateTask(task *Task) error {
	query := `
		UPDATE scheduler 
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?
	`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func defaultTasks(limit int) ([]*Task, error) {
	rows, err := DB.Query(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		ORDER BY date
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, nil
}
