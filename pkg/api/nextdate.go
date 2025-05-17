package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const dateFormat = "20060102"

// afterNow сравнивает две даты (без времени).
func afterNow(a, b time.Time) bool {
	a = a.Truncate(24 * time.Hour)
	b = b.Truncate(24 * time.Hour)
	return a.After(b)
}

// lastDayOfMonth возвращает последний день месяца у даты t.
func lastDayOfMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

// NextDate вычисляет следующую дату выполнения задачи.
// now    — точка отсчёта,
// dstart — исходная дата в формате YYYYMMDD,
// repeat — правило повторения ("y", "d N", "w ...", "m ...").
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// парсим стартовую дату
	start, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("неверная дата начала: %w", err)
	}
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не задано")
	}

	parts := strings.Fields(repeat)
	switch parts[0] {
	case "y":
		// каждый год
		for {
			start = start.AddDate(1, 0, 0)
			if afterNow(start, now) {
				break
			}
		}

	case "d":
		// через N дней
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат правила d")
		}
		var interval int
		_, err := fmt.Sscanf(parts[1], "%d", &interval)
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("интервал d должен быть от 1 до 400")
		}
		for {
			start = start.AddDate(0, 0, interval)
			if afterNow(start, now) {
				break
			}
		}

	case "w":
		// по дням недели
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат правила w")
		}
		days := strings.Split(parts[1], ",")
		valid := map[time.Weekday]bool{}
		for _, tok := range days {
			var d int
			_, err := fmt.Sscanf(tok, "%d", &d)
			if err != nil || d < 1 || d > 7 {
				return "", fmt.Errorf("некорректный день недели: %s", tok)
			}
			// в Go Weekday: 0=Sunday; нам 1=Monday → сместим
			valid[time.Weekday((d+6)%7)] = true
		}
		// ищем следующий день
		cur := start
		for {
			if afterNow(cur, now) && valid[cur.Weekday()] {
				start = cur
				break
			}
			cur = cur.AddDate(0, 0, 1)
		}

	case "m":
		// по дням месяца и (опционально) месяцам
		if len(parts) < 2 {
			return "", fmt.Errorf("некорректный формат правила m")
		}
		daySet := make(map[int]bool)
		for _, tok := range strings.Split(parts[1], ",") {
			var dd int
			_, err := fmt.Sscanf(tok, "%d", &dd)
			if err != nil || dd == 0 || dd < -2 || dd > 31 {
				return "", fmt.Errorf("некорректный день месяца: %s", tok)
			}
			daySet[dd] = true
		}
		monthSet := make(map[time.Month]bool)
		if len(parts) > 2 {
			for _, tok := range strings.Split(parts[2], ",") {
				var mm int
				_, err := fmt.Sscanf(tok, "%d", &mm)
				if err != nil || mm < 1 || mm > 12 {
					return "", fmt.Errorf("некорректный месяц: %s", tok)
				}
				monthSet[time.Month(mm)] = true
			}
		}
		cur := start
		for {
			if afterNow(cur, now) {
				// если указаны месяцы и текущий не подходит — пропускаем
				if len(monthSet) > 0 && !monthSet[cur.Month()] {
					cur = cur.AddDate(0, 0, 1)
					continue
				}
				day := cur.Day()
				last := lastDayOfMonth(cur)
				// проверяем обычные числа, последний и предпоследний
				if daySet[day] || (daySet[-1] && day == last) || (daySet[-2] && day == last-1) {
					start = cur
					break
				}
			}
			cur = cur.AddDate(0, 0, 1)
		}

	default:
		return "", fmt.Errorf("неподдерживаемый формат повторения")
	}

	return start.Format(dateFormat), nil
}

// nextDateHandler обрабатывает GET /api/nextdate и возвращает просто строку даты.
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// парсим параметры
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// готовим now
	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "неправильный формат now", http.StatusBadRequest)
			return
		}
	}

	// вычисляем next
	next, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// возвращаем плейн-текстом
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(next))
}
