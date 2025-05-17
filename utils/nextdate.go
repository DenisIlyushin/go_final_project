package utils

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"slices"

	"github.com/DenisIlyushin/go_final_project/config"
)

var (
	ErrEmptyRepeat          = errors.New("правило повторения не задано")
	ErrDaysRange            = errors.New("количество дней должно быть от 1 до 400")
	ErrInvalidRepeatFormatW = errors.New("неверный формат недельного правила")
	ErrWeekdayRange         = errors.New("день недели должен быть от 1 до 7")
	ErrMonthsFormat         = errors.New("неверный формат месячного правила")
	ErrDaysOfMonthRange     = errors.New("день месяца должен быть 1–31 или -1/-2")
	ErrMonthsRange          = errors.New("месяц должен быть от 1 до 12")
	ErrRepeatRuleNotFound   = errors.New("правило повторения не найдено")
)

// NextDate вычисляет следующую дату задачи на основе текущей даты now,
// стартовой даты date и правила repeat.
// Поддерживаемые правила:
//   - "d N": каждые N дней (1 ≤ N ≤ 400);
//   - "y": ежегодно;
//   - "w d1,d2,...": еженедельно по дням недели (1–7);
//   - "m D[,...][ M[,...]]": ежемесячно по числам D (1–31, -1 последний, -2 предпоследний),
//     опционально ограничено месяцами M (1–12).
func NextDate(now time.Time, date, repeat string) (string, error) {
	taskDate, err := parseDate(date)
	if err != nil {
		return "", err
	}
	if repeat == "" {
		return "", ErrEmptyRepeat
	}
	now = now.Truncate(24 * time.Hour)

	parts := strings.Fields(repeat)
	rule := parts[0]
	param := ""
	if len(parts) > 1 {
		param = strings.Join(parts[1:], " ")
	}

	switch rule {
	case "d":
		return nextDaily(taskDate, now, param)
	case "y":
		return nextYearly(taskDate, now), nil
	case "w":
		return nextWeekly(taskDate, now, param)
	case "m":
		return nextMonthly(taskDate, now, param)
	default:
		return "", ErrRepeatRuleNotFound
	}
}

// parseDate парсит строку date в формате config.DateFormat.
func parseDate(date string) (time.Time, error) {
	return time.Parse(config.DateFormat, date)
}

// nextDaily вычисляет следующий день с шагом days.
func nextDaily(taskDate, now time.Time, arg string) (string, error) {
	days, err := strconv.Atoi(arg)
	if err != nil {
		return "", err
	}
	if days < 1 || days > 400 {
		return "", ErrDaysRange
	}
	next := taskDate.AddDate(0, 0, days)
	for !next.After(now) {
		next = next.AddDate(0, 0, days)
	}
	return next.Format(config.DateFormat), nil
}

// nextYearly вычисляет следующую дату с ежегодным шагом.
func nextYearly(taskDate, now time.Time) string {
	next := taskDate.AddDate(1, 0, 0)
	for !next.After(now) {
		next = next.AddDate(1, 0, 0)
	}
	return next.Format(config.DateFormat)
}

// nextWeekly вычисляет следующую дату по дням недели из param (e.g. "1,3,5").
func nextWeekly(taskDate, now time.Time, param string) (string, error) {
	if param == "" {
		return "", ErrInvalidRepeatFormatW
	}
	parts := strings.Split(param, ",")
	days := make([]int, 0, len(parts))
	for _, p := range parts {
		w, err := strconv.Atoi(p)
		if err != nil {
			return "", err
		}
		if w < 1 || w > 7 {
			return "", ErrWeekdayRange
		}
		days = append(days, w)
	}
	sort.Ints(days)

	base := now
	if now.Before(taskDate) {
		base = taskDate
	}
	for _, w := range days {
		delta := (w - int(base.Weekday()) + 7) % 7
		if delta == 0 {
			delta = 7
		}
		next := base.AddDate(0, 0, delta)
		if next.After(now) {
			return next.Format(config.DateFormat), nil
		}
	}
	// переходим к первому дню в следующей неделе
	delta := (days[0] - int(base.Weekday()) + 7) % 7
	if delta == 0 {
		delta = 7
	}
	next := base.AddDate(0, 0, delta)
	return next.Format(config.DateFormat), nil
}

// nextMonthly вычисляет следующую дату по правилам месячного повторения.
func nextMonthly(taskDate, now time.Time, param string) (string, error) {
	parts := strings.Fields(param)
	if len(parts) < 1 || len(parts) > 2 {
		return "", ErrMonthsFormat
	}
	// дни месяца
	dayParts := strings.Split(parts[0], ",")
	days := make([]int, 0, len(dayParts))
	for _, d := range dayParts {
		n, err := strconv.Atoi(d)
		if err != nil {
			return "", err
		}
		if !((n > 0 && n < 32) || n == -1 || n == -2) {
			return "", ErrDaysOfMonthRange
		}
		days = append(days, n)
	}
	sort.Ints(days)
	// месяцы (опционально)
	months := make([]int, 0)
	if len(parts) == 2 {
		for _, m := range strings.Split(parts[1], ",") {
			n, err := strconv.Atoi(m)
			if err != nil {
				return "", err
			}
			if n < 1 || n > 12 {
				return "", ErrMonthsRange
			}
			months = append(months, n)
		}
		sort.Ints(months)
	}
	base := now
	if now.Before(taskDate) {
		base = taskDate
	}
	for {
		base = base.AddDate(0, 0, 1)
		if len(months) > 0 && !slices.Contains(months, int(base.Month())) {
			base = time.Date(base.Year(), base.Month()+1, 1, 0, 0, 0, 0, base.Location())
			continue
		}
		if slices.Contains(days, base.Day()) {
			if base.After(now) {
				return base.Format(config.DateFormat), nil
			}
		}
		// проверка последнего/предпоследнего дня
		y, mm, _ := base.Date()
		lastDay := time.Date(y, mm+1, 0, 0, 0, 0, 0, base.Location()).Day()
		if (slices.Contains(days, -1) && base.Day() == lastDay) ||
			(slices.Contains(days, -2) && base.Day() == lastDay-1) {
			if base.After(now) {
				return base.Format(config.DateFormat), nil
			}
		}
	}
}
