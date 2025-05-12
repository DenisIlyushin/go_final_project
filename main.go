package main

import (
	"diploma/pkg/server"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// Загружаем переменные окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Нет .env файла, использую переменные окружения из системы")
	}

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
