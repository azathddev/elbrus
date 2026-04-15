package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gigachat "github.com/tigusigalpa/gigachat-go"
)

type generateRequest struct {
	Message  string `json:"message"`
	Audience string `json:"audience"`
}

type generateResponse struct {
	ImageURL    string `json:"imageUrl,omitempty"`
	ImageBase64 string `json:"imageBase64,omitempty"`
	Prompt      string `json:"prompt"`
}

func main() {
	authKey := resolveAuthKey()
	model := strings.TrimSpace(os.Getenv("GIGACHAT_IMAGE_MODEL"))
	if model == "" {
		model = gigachat.GigaChat2Max
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8081"
	}

	tokenManager := gigachat.NewTokenManager(authKey)
	client := gigachat.NewClient(
		tokenManager,
		gigachat.WithDefaultModel(model),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}

		req.Message = strings.TrimSpace(req.Message)
		req.Audience = strings.TrimSpace(req.Audience)
		if err := validateRequest(req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		prompt := buildPrompt(req.Message, req.Audience)
		imageResp, err := client.CreateImage(
			prompt,
			gigachat.WithImageModel(model),
		)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if strings.TrimSpace(imageResp.Content) == "" {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "SDK returned empty image response"})
			return
		}

		result := generateResponse{
			ImageURL:    "",
			ImageBase64: imageResp.Content,
			Prompt:      prompt,
		}

		writeJSON(w, http.StatusOK, map[string]any{"result": result})
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("SDK backend started on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func resolveAuthKey() string {
	authKey := strings.TrimSpace(os.Getenv("GIGACHAT_AUTH_KEY"))
	if authKey == "" {
		// Backward-compatibility for previous variable naming.
		authKey = strings.TrimSpace(os.Getenv("GIGACHAT_AUTH_TOKEN"))
	}
	if authKey != "" {
		return authKey
	}

	clientID := strings.TrimSpace(os.Getenv("GIGACHAT_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GIGACHAT_CLIENT_SECRET"))
	if clientID != "" && clientSecret != "" {
		return base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}

	log.Fatal("set GIGACHAT_AUTH_KEY (or GIGACHAT_AUTH_TOKEN), or GIGACHAT_CLIENT_ID + GIGACHAT_CLIENT_SECRET")
	return ""
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func validateRequest(req generateRequest) error {
	if req.Message == "" {
		return errors.New("message is required")
	}
	if req.Audience == "" {
		return errors.New("audience is required")
	}
	return nil
}

func buildPrompt(message, audience string) string {
	return "Нарисуй рекламный баннер для аудитории \"" + audience + "\" по сообщению: \"" + message + "\". Стиль: современный, чистый, контрастный, подходит для сайта и соцсетей."
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
