package handlers

import (
	"fmt"
	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/database"
	"github.com/DenisIlyushin/go_final_project/utils"
	"net/http"
	"time"
)

type TaskHandler struct {
	DB       *database.Database
	Settings *config.Settings
}

func NewTaskHandler(database *database.Database, settings *config.Settings) *TaskHandler {
	return &TaskHandler{
		DB:       database,
		Settings: settings,
	}
}

// NextDate обрабатывает запрос и возвращает следующую дату выполнения задачи.
// Ожидаемые параметры формы:
//   - now: текущая дата в формате YYYYMMDD
//   - date: стартовая дата задачи в формате YYYYMMDD
//   - repeat: правило повторения (см. utils.NextDate)
//
// В ответ возвращается JSON {"nextDate":"YYYYMMDD"} или {"error":"..."}.
func (h *TaskHandler) NextDate(w http.ResponseWriter, r *http.Request) {
	timeNow := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if timeNow == "" || date == "" || repeat == "" {
		http.Error(w, `{"error":"missing some parameters"}`, http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", timeNow)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	_, err = time.Parse("20060102", date)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	nextDate, err := utils.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(nextDate)); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
}
