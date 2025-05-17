package server

import (
	"github.com/DenisIlyushin/go_final_project/handlers"
	"log"
	"net/http"

	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/database"
)

func Run(database *database.Database, settings *config.Settings) error {

	taskHandler := handlers.NewTaskHandler(database, settings)
	router := serveTaskRouter(taskHandler)

	log.Println("Сервер запущен на порту:", settings.ServerPort)
	return http.ListenAndServe(":"+settings.ServerPort, router)
}
