package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/models"
)

var (
	ErrEmptyTitle          = errors.New("название задачи не может быть пустым")
	ErrInvalidDateFormat   = errors.New("неверный формат даты")
	ErrNextDateCalculation = errors.New("ошибка вычисления следующей даты")
)

// ValidateTask проверяет и нормализует поля задачи t перед её созданием или обновлением.
// Параметр now задаёт текущую дату в формате config.DateFormat.
func ValidateTask(t *models.Task, now string) error {
	// Проверяем название
	if t.Title == "" {
		return ErrEmptyTitle
	}

	// Устанавливаем дату, если не задана
	if t.Date == "" {
		t.Date = now
	}

	// Валидация формата даты
	if _, err := time.Parse(config.DateFormat, t.Date); err != nil {
		return ErrInvalidDateFormat
	}

	// Если дата раньше текущей
	if t.Date < now {
		if t.Repeat == "" {
			t.Date = now
		} else {
			// Вычисляем следующую дату по правилу Repeat
			next, err := NextDate(time.Now(), t.Date, t.Repeat)
			if err != nil {
				return fmt.Errorf("%s: %w", ErrNextDateCalculation, err)
			}
			t.Date = next
		}
	}
	return nil
}
