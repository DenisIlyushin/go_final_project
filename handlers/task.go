package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/database"
	"github.com/DenisIlyushin/go_final_project/models"
	"github.com/DenisIlyushin/go_final_project/utils"
)

const (
	ErrMissingParams  = "не заданы необходимые параметры"
	ErrInvalidNowDate = "некорректный формат текущей даты"
	ErrInvalidDate    = "некорректный формат даты задачи"
	ErrDecodeBody     = "не удалось декодировать тело запроса"
	ErrEmptyTitle     = "название задачи не может быть пустым"
	ErrInternalCreate = "внутренняя ошибка сервера при создании задачи"
)

// TaskHandler обрабатывает HTTP-запросы, связанные с задачами.
type TaskHandler struct {
	DB       *database.Database
	Settings *config.Settings
}

// NewTaskHandler создает новый обработчик задач.
func NewTaskHandler(db *database.Database, settings *config.Settings) *TaskHandler {
	return &TaskHandler{DB: db, Settings: settings}
}

// NextDate обрабатывает GET-запрос и возвращает следующую дату выполнения задачи.
func (h *TaskHandler) NextDate(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	if nowParam == "" || dateParam == "" || repeatParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrMissingParams})
		return
	}
	parseDate := func(val string) (time.Time, error) {
		return time.Parse(config.DateFormat, val)
	}
	nowTime, err := parseDate(nowParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidNowDate})
		return
	}
	if _, err := parseDate(dateParam); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidDate})
		return
	}
	next, err := utils.NextDate(nowTime, dateParam, repeatParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"nextDate": next})
}

// CreateTask обрабатывает POST-запрос на создание новой задачи.
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrDecodeBody})
		return
	}
	if strings.TrimSpace(task.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrEmptyTitle})
		return
	}
	now := time.Now().Truncate(24 * time.Hour)
	if task.Date == "" || strings.EqualFold(task.Date, "today") {
		task.Date = now.Format(config.DateFormat)
	}
	parsedDate, err := time.Parse(config.DateFormat, task.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidDate})
		return
	}
	// Если дата раньше текущей
	if parsedDate.Before(now) {
		if task.Repeat == "" {
			// устанавливаем на сегодня
			task.Date = now.Format(config.DateFormat)
		} else {
			next, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			task.Date = next
		}
	}

	id, err := h.DB.CreateTask(task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrInternalCreate})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// writeJSON устанавливает заголовок Content-Type, статус и кодирует data в JSON.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
