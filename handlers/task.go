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
	ErrMissingParams    = "не заданы необходимые параметры"
	ErrInvalidNowDate   = "некорректный формат текущей даты"
	ErrInvalidDate      = "некорректный формат даты задачи"
	ErrInvalidJSON      = "неверный формат JSON"
	ErrDecodeBody       = "не удалось декодировать тело запроса"
	ErrEmptyTitle       = "название задачи не может быть пустым"
	ErrInternalCreate   = "не удалось создать задачу"
	ErrMethodNotAllowed = "метод не поддерживается"
	ErrGetTasksDB       = "не удалось получить список задач"
	ErrGetTaskDB        = "не удалось получить задачу"
	ErrEncodeResponse   = "не удалось сформировать ответ"
	ErrDBUpdate         = "не удалось обновить задачу"
	ErrMissingID        = "не указан идентификатор задачи"
	ErrDBDelete         = "не удалось удалить задачу"
	ErrDBDone           = "не удалось отметить задачу выполненной"
)

// TaskHandler обрабатывает HTTP-запросы, связанные с задачами.
type TaskHandler struct {
	DB       *database.Database
	Settings *config.Settings
}

// NewTaskHandler создаёт новый экземпляр TaskHandler с заданным подключением к БД и настройками.
func NewTaskHandler(db *database.Database, settings *config.Settings) *TaskHandler {
	return &TaskHandler{DB: db, Settings: settings}
}

// NextDate обрабатывает GET-запрос и возвращает следующую дату выполнения задачи.
// Параметры запроса:
// - now: текущая дата в формате YYYYMMDD
// - date: исходная дата задачи в формате YYYYMMDD
// - repeat: правило повторения
// В ответ возвращается JSON {"nextDate":"YYYYMMDD"} или {"error":"..."}.
func (h *TaskHandler) NextDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	if nowParam == "" || dateParam == "" || repeatParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrMissingParams})
		return
	}
	parse := func(val string) (time.Time, error) {
		return time.Parse(config.DateFormat, val)
	}
	nowTime, err := parse(nowParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidNowDate})
		return
	}
	if _, err := parse(dateParam); err != nil {
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
// В теле запроса ожидается JSON с полями модели Task.
// В ответ возвращается JSON {"id":<ID>} или {"error":"..."}.
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
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
	if _, err := time.Parse(config.DateFormat, task.Date); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidDate})
		return
	}
	parsed, _ := time.Parse(config.DateFormat, task.Date)
	if parsed.Before(now) {
		if task.Repeat == "" {
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

// GetTasks обрабатывает GET-запрос и возвращает список задач.
// Опциональный параметр search фильтрует задачи по дате (формат YYYYMMDD или DD.MM.YYYY)
// или по вхождению в title/comment.
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	search := r.URL.Query().Get("search")
	if search != "" {
		if t, err := time.Parse("02.01.2006", search); err == nil {
			search = t.Format(config.DateFormat)
		}
	}
	tasks, err := h.DB.GetTasks(search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrGetTasksDB})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// GetTask обрабатывает GET-запрос и возвращает задачу по её ID.
// Ожидается параметр id в query.
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	id := r.URL.Query().Get("id")
	task, err := h.DB.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrGetTaskDB})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// UpdateTask обрабатывает PUT /api/task и обновляет задачу.
// В теле JSON c полями модели Task (включая id).
// В ответе пустой JSON {} или {"error":"..."}.
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrInvalidJSON})
		return
	}
	now := time.Now().Format(config.DateFormat)
	if err := utils.ValidateTask(&t, now); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.DB.EditTask(t); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBUpdate})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

// DeleteTask обрабатывает DELETE /api/task?id=<ID>.
// В случае успеха возвращает {}, иначе {"error":"…"}.
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrMissingID})
		return
	}
	if err := h.DB.DeleteTask(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBDelete})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

// CompleteTask обрабатывает POST /api/task/done?id=<ID>:
// • если task.Repeat == "", удаляет задачу;
// • иначе вычисляет новую дату через utils.NextDate и вызывает EditTask.
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": ErrMethodNotAllowed})
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ErrMissingID})
		return
	}
	task, err := h.DB.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBDone})
		return
	}
	if task.Repeat == "" {
		// одноразовая – удаляем
		if err := h.DB.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBDone})
			return
		}
	} else {
		// периодическая – вычисляем дату и обновляем
		next, err := utils.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		task.Date = next
		if err := h.DB.EditTask(task); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBDone})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

// writeJSON устанавливает заголовок Content-Type, статус и кодирует ответ в JSON.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, ErrEncodeResponse, http.StatusInternalServerError)
	}
}
