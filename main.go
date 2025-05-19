package main

import (
	"log"

	"github.com/DenisIlyushin/go_final_project/config"
	"github.com/DenisIlyushin/go_final_project/database"
	"github.com/DenisIlyushin/go_final_project/server"
)

func main() {
	// Применяем конфиг
	config := config.LoadConfig()

	// Инициализируем базу
	db, err := database.OpenDatabase(config.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем сервер
	if err := server.Run(db, config); err != nil {
		log.Fatal(err)
	}
}
