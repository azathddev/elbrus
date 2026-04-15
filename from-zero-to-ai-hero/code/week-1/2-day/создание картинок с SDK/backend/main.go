package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	insecureSkipVerify, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("GIGACHAT_INSECURE_SKIP_VERIFY")))
	customCAPath := strings.TrimSpace(os.Getenv("GIGACHAT_CA_CERT_PATH"))
	httpClient, err := buildHTTPClient(insecureSkipVerify, customCAPath)
	if err != nil {
		log.Fatalf("configure TLS: %v", err)
	}
	if insecureSkipVerify {
		log.Println("WARNING: GIGACHAT_INSECURE_SKIP_VERIFY=true disables TLS certificate verification")
	}

	tokenManager := gigachat.NewTokenManager(
		authKey,
		gigachat.WithTokenManagerHTTPClient(httpClient),
	)
	client := gigachat.NewClient(
		tokenManager,
		gigachat.WithDefaultModel(model),
		gigachat.WithHTTPClient(httpClient),
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

func buildHTTPClient(insecureSkipVerify bool, customCAPath string) (*http.Client, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: insecureSkipVerify,
	}

	if customCAPath != "" {
		caPEM, err := os.ReadFile(customCAPath)
		if err != nil {
			return nil, fmt.Errorf("read custom CA file %q: %w", customCAPath, err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil || roots == nil {
			roots = x509.NewCertPool()
		}
		if err := appendCustomCA(roots, caPEM, customCAPath); err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = roots
		log.Printf("custom CA loaded from %s", customCAPath)
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
		TLSClientConfig:       tlsConfig,
	}

	return &http.Client{
		Timeout:   120 * time.Second,
		Transport: transport,
	}, nil
}

func appendCustomCA(pool *x509.CertPool, certData []byte, sourcePath string) error {
	if ok := pool.AppendCertsFromPEM(certData); ok {
		return nil
	}

	cert, err := x509.ParseCertificate(certData)
	if err == nil {
		pool.AddCert(cert)
		return nil
	}

	return fmt.Errorf(
		"failed to parse CA certificate %q as PEM or DER: %w",
		sourcePath,
		err,
	)
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
