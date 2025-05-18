package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/DenisIlyushin/go_final_project/auth"
	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/handlers"
)

func serveTaskRouter(handler *handlers.TaskHandler, settings *config.Settings) *chi.Mux {
	// Создаём сервис аутентификации
	authService := auth.NewService(settings)

	router := chi.NewRouter()

	router.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(config.FrontendPath))))

	router.Post("/api/signin", authService.Signin)
	router.Get("/api/nextdate", handler.NextDate)

	router.Route("/api", func(r chi.Router) {
		r.Use(authService.Middleware)
		r.Post("/task", handler.CreateTask)
		r.Get("/tasks", handler.GetTasks)
		r.Get("/task", handler.GetTask)
		r.Put("/task", handler.UpdateTask)
		r.Post("/task/done", handler.CompleteTask)
		r.Delete("/task", handler.DeleteTask)
	})

	return router
}
