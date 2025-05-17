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

	//router.Get("/api/nextdate", handler.NextDateHandler)
	//
	//router.Post("/api/task", handler.PostTask)
	//router.Get("/api/tasks", handler.GetTasksHandler)
	//router.Get("/api/task", handler.GetTaskHandler)
	//router.Post("/api/task/done", handler.DoneTaskHandler)
	//router.Put("/api/task", handler.UpdateTaskHandler)
	//router.Delete("/api/task", handler.DeleteTaskHandler)

	return router
}
