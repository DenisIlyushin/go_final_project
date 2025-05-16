package api

import (
	"diploma/pkg/db"
	"net/http"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	if tasks == nil {
		tasks = []*db.Task{}
	}
	writeJson(w, TasksResp{Tasks: tasks})
}
