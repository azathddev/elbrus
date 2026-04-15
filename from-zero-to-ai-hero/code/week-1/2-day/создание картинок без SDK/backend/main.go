package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	api "banner-generator/backend/internal/http"
	"banner-generator/backend/internal/service"
)

func main() {
	token := strings.TrimSpace(os.Getenv("GIGACHAT_AUTH_TOKEN"))
	if token == "" {
		// Compatibility with SDK naming style.
		token = strings.TrimSpace(os.Getenv("GIGACHAT_AUTH_KEY"))
	}
	if token == "" {
		log.Fatal("GIGACHAT_AUTH_TOKEN (or GIGACHAT_AUTH_KEY) is required")
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
	insecureSkipVerify, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("GIGACHAT_INSECURE_SKIP_VERIFY")))
	if insecureSkipVerify {
		log.Println("WARNING: GIGACHAT_INSECURE_SKIP_VERIFY=true disables TLS certificate verification")
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   20 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecureSkipVerify,
		},
	}
	client := &http.Client{
		Timeout:   120 * time.Second,
		Transport: transport,
	}
	generator := service.NewImageGenerator(client, apiURL, token, model)
	handler := api.NewHandler(generator)

	log.Printf("server started on :%s", port)
	if err := http.ListenAndServe(":"+port, api.NewRouter(handler)); err != nil {
		log.Fatal(err)
	}
}
