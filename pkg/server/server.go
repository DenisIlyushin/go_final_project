package server

import (
	"diploma/pkg/api"
	"fmt"
	"net/http"
	"os"
)

func Run() error {
	webDir := "web"
	port := "7540"

	api.Init()

	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	fmt.Println("Сервер запущен на порту:", port)
	return http.ListenAndServe(":"+port, nil)
}
