package db

import (
	"database/sql"
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

func Tasks(limit int, search string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		if t, errDate := time.Parse("02.01.2006", search); errDate == nil {
			// Поиск по дате
			date := t.Format("20060102")
			rows, err = DB.Query(`
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE date = ?
				ORDER BY date
				LIMIT ?`, date, limit)
		} else {
			// Поиск по словам в title и comment, нечувствительно к регистру
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
