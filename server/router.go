package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/handlers"
)

func serveTaskRouter(handler *handlers.TaskHandler) *chi.Mux {
	router := chi.NewRouter()

	router.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(config.FrontendPath))))

	router.Get("/api/nextdate", handler.NextDate)

	router.Post("/api/task", handler.CreateTask)
	router.Get("/api/tasks", handler.GetTasks)
	router.Get("/api/task", handler.GetTask)
	//router.Put("/api/task", handler.UpdateTaskHandler)
	//router.Post("/api/task/done", handler.DoneTaskHandler)
	//router.Delete("/api/task", handler.DeleteTaskHandler)

	return router
}
