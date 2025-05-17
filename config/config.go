package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	noEnvWarning  = "no .env file found, using system environment"
	noPortWarning = "TODO_PORT not set, using default"
	noPathWarning = "TODO_DBFILE not set, using default"
)

type Settings struct {
	ServerPort   string
	DatabasePath string
}

// NewConfig загружает переменные окружения из файла .env (если он существует)
// или устанавливает значения по умолчанию.
// Возвращает указатель на экземпляр Settings с готовыми к использованию значениями.
func LoadConfig() *Settings {
	if err := godotenv.Load(); err != nil {
		log.Println(noEnvWarning)
	}

	return &Settings{
		ServerPort:   getenv("TODO_PORT", DefaultPort, noPortWarning),
		DatabasePath: getenv("TODO_DBFILE", DefaultDatabasePath, noPathWarning),
	}
}

// getenv читает переменную окружения key и возвращает её значение,
// либо fallback и лог-месседж при пустом результате.
func getenv(key, fallback, warnMsg string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	log.Println(warnMsg)
	return fallback
}
