package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const dateFormat = "20060102"

func afterNow(a, b time.Time) bool {
	a = a.Truncate(24 * time.Hour)
	b = b.Truncate(24 * time.Hour)
	return a.After(b)
}

func lastDayOfMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
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
		for {
			start = start.AddDate(1, 0, 0)
			if afterNow(start, now) {
				break
			}
		}

	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат правила d")
		}
		var interval int
		_, err := fmt.Sscanf(parts[1], "%d", &interval)
		if err != nil || interval <= 0 || interval > 400 {
			return "", fmt.Errorf("интервал d должен быть от 1 до 400")
		}
		for {
			start = start.AddDate(0, 0, interval)
			if afterNow(start, now) {
				break
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("неправильный формат w")
		}
		days := strings.Split(parts[1], ",")
		valid := map[time.Weekday]bool{}
		for _, d := range days {
			var n int
			_, err := fmt.Sscanf(d, "%d", &n)
			if err != nil || n < 1 || n > 7 {
				return "", fmt.Errorf("некорректный день недели: %s", d)
			}
			valid[time.Weekday((n+6)%7)] = true // Go: 0=Sunday, а у нас 1=Monday
		}
		date := start
		for {
			if afterNow(date, now) && valid[date.Weekday()] {
				break
			}
			date = date.AddDate(0, 0, 1)
		}
		return date.Format(dateFormat), nil

	case "m":
		if len(parts) < 2 {
			return "", fmt.Errorf("неправильный формат m")
		}

		daySet := make(map[int]bool)
		for _, s := range strings.Split(parts[1], ",") {
			var d int
			_, err := fmt.Sscanf(s, "%d", &d)
			if err != nil || d == 0 || d < -2 || d > 31 {
				return "", fmt.Errorf("некорректный день месяца: %s", s)
			}
			daySet[d] = true
		}

		monthSet := make(map[time.Month]bool)
		if len(parts) > 2 {
			for _, s := range strings.Split(parts[2], ",") {
				var m int
				_, err := fmt.Sscanf(s, "%d", &m)
				if err != nil || m < 1 || m > 12 {
					return "", fmt.Errorf("некорректный месяц: %s", s)
				}
				monthSet[time.Month(m)] = true
			}
		}

		date := start
		for {
			if afterNow(date, now) {
				m := date.Month()
				if len(monthSet) > 0 && !monthSet[m] {
					goto nextDay
				}
				day := date.Day()
				last := lastDayOfMonth(date)

				if daySet[day] ||
					(daySet[-1] && day == last) ||
					(daySet[-2] && day == last-1) {
					break
				}
			}
		nextDay:
			date = date.AddDate(0, 0, 1)
		}
		return date.Format(dateFormat), nil

	default:
		return "", fmt.Errorf("неподдерживаемый формат повторения")
	}

	return start.Format(dateFormat), nil
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{
				"error": "неправильный формат now",
			})
			return
		}
	}

	next, err := NextDate(now, date, repeat)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"date": next,
	})
}
