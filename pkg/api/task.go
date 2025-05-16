package api

import (
	"diploma/pkg/db"
	"encoding/json"
	"net/http"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "Не указан id")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJson(w, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "Неверный JSON")
		return
	}

	if task.ID == "" || task.Title == "" {
		writeError(w, "Отсутствует id или title")
		return
	}

	if err := checkDate(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	writeJson(w, map[string]string{})
}
