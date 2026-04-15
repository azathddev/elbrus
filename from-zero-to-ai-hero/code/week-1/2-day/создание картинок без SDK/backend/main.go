package main

import (
	"log"
	"net/http"
	"os"
	"time"

	api "banner-generator/backend/internal/http"
	"banner-generator/backend/internal/service"
)

func main() {
	token := os.Getenv("GIGACHAT_AUTH_TOKEN")
	if token == "" {
		log.Fatal("GIGACHAT_AUTH_TOKEN is required")
	}

	apiURL := os.Getenv("GIGACHAT_IMAGE_API_URL")
	if apiURL == "" {
		apiURL = "https://gigachat.devices.sberbank.ru/api/v1/images/generations"
	}

	model := os.Getenv("GIGACHAT_IMAGE_MODEL")
	if model == "" {
		model = "GigaChat-2-Max"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{Timeout: 90 * time.Second}
	generator := service.NewImageGenerator(client, apiURL, token, model)
	handler := api.NewHandler(generator)

	log.Printf("server started on :%s", port)
	if err := http.ListenAndServe(":"+port, api.NewRouter(handler)); err != nil {
		log.Fatal(err)
	}
}
