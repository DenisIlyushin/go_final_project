package main

import (
	"diploma/pkg/db"
	"diploma/pkg/server"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	// Загружаем переменные окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла, использую переменные окружения из системы")
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
