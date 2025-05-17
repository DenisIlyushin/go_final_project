package handlers

import (
	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/database"
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
