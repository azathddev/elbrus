package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"banner-generator/backend/internal/service"
)

type Handler struct {
	generator *service.ImageGenerator
}

type generateRequest struct {
	Message  string `json:"message"`
	Audience string `json:"audience"`
}

type apiError struct {
	Error string `json:"error"`
}

func NewHandler(generator *service.ImageGenerator) *Handler {
	return &Handler{generator: generator}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid JSON body"})
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	req.Audience = strings.TrimSpace(req.Audience)
	if err := validateGenerateRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	result, err := h.generator.Generate(r.Context(), service.GenerateRequest{
		Message:  req.Message,
		Audience: req.Audience,
	})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, apiError{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func validateGenerateRequest(req generateRequest) error {
	if req.Message == "" {
		return errors.New("message is required")
	}
	if req.Audience == "" {
		return errors.New("audience is required")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
