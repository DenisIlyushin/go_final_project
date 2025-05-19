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
	ErrMissingParams    = "обязательные параметры запроса не заданы: now, date, repeat"
	ErrInvalidNowDate   = "текущая дата имеет неверный формат, ожидается YYYYMMDD"
	ErrInvalidDate      = "дата задачи имеет неверный формат, ожидается YYYYMMDD"
	ErrInvalidJSON      = "некорректный JSON в теле запроса"
	ErrDecodeBody       = "ошибка декодирования тела запроса"
	ErrEmptyTitle       = "поле Title не может быть пустым"
	ErrInternalCreate   = "ошибка при создании задачи"
	ErrMethodNotAllowed = "метод запроса не поддерживается"
	ErrGetTasksDB       = "ошибка при получении списка задач"
	ErrGetTaskDB        = "ошибка при получении задачи"
	ErrEncodeResponse   = "ошибка при формировании JSON-ответа"
	ErrDBUpdate         = "ошибка при обновлении задачи"
	ErrMissingID        = "не указан идентификатор задачи"
	ErrDBDelete         = "ошибка при удалении задачи"
	ErrDBDone           = "ошибка при завершении задачи"
)

// TaskHandler обрабатывает HTTP-запросы, связанные с задачами.
// DB — подключение к базе данных, Settings — настройки приложения.
type TaskHandler struct {
	DB       *database.Database
	Settings *config.Settings
}

// NewTaskHandler создаёт TaskHandler с указанными подключением к БД и настройками.
func NewTaskHandler(db *database.Database, settings *config.Settings) *TaskHandler {
	return &TaskHandler{DB: db, Settings: settings}
}

// NextDate возвращает следующую дату выполнения задачи в формате YYYYMMDD.
// Ожидает GET-параметры now (текущая дата YYYYMMDD), date (дата задачи YYYYMMDD), repeat (правило повторения).
// При ошибке возвращает JSON {"error": "..."} и соответствующий HTTP статус.
// При успешном вычислении возвращает plain text с датой и статус 200.
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
	// возвращаем только строку даты
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(next))
}

// CreateTask создаёт новую задачу.
// Ожидает POST с JSON модели Task. В ответе JSON {"id":ID} или {"error":"..."}.
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

// GetTasks возвращает список задач. Опционально по search фильтрует по дате (YYYYMMDD или DD.MM.YYYY) или вхождению текста.
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

// GetTask возвращает задачу по её идентификатору из query id.
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

// UpdateTask обновляет существующую задачу.
// Ожидает PUT с JSON модели Task (включая id). Возвращает {} или {"error":"..."}.
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

// DeleteTask удаляет задачу по id из query. Возвращает {} или {"error":"..."}.
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

// CompleteTask отмечает задачу выполненной: одноразовые удаляются, периодические получают новую дату.
// Ожидает POST /api/task/done?id=<ID>.
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
		if err := h.DB.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": ErrDBDone})
			return
		}
	} else {
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

// writeJSON устанавливает Content-Type, статус и кодирует data в JSON.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, ErrEncodeResponse, http.StatusInternalServerError)
	}
}
