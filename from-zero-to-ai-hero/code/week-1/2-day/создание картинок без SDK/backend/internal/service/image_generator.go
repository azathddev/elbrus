package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type ImageGenerator struct {
	client *http.Client
	apiURL string
	token  string
	model  string
}

type GenerateRequest struct {
	Message  string `json:"message"`
	Audience string `json:"audience"`
}

type GenerateResult struct {
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageBase string `json:"imageBase64,omitempty"`
	Prompt    string `json:"prompt"`
}

type generatorRequest struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
}

func NewImageGenerator(client *http.Client, apiURL, token, model string) *ImageGenerator {
	return &ImageGenerator{
		client: client,
		apiURL: apiURL,
		token:  token,
		model:  model,
	}
}

func (g *ImageGenerator) Generate(ctx context.Context, req GenerateRequest) (GenerateResult, error) {
	prompt := buildPrompt(req.Message, req.Audience)
	payload, err := json.Marshal(generatorRequest{
		Model:  g.model,
		Prompt: prompt,
	})
	if err != nil {
		return GenerateResult{}, fmt.Errorf("marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apiURL, bytes.NewReader(payload))
	if err != nil {
		return GenerateResult{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+g.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := g.doWithRetry(httpReq, 3)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("call image API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return GenerateResult{}, fmt.Errorf("image API failed with status %d: %v", resp.StatusCode, body)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return GenerateResult{}, fmt.Errorf("decode API response: %w", err)
	}

	result, err := parseImageResult(raw)
	if err != nil {
		return GenerateResult{}, err
	}
	result.Prompt = prompt

	return result, nil
}

func (g *ImageGenerator) doWithRetry(req *http.Request, attempts int) (*http.Response, error) {
	var lastErr error
	for i := 1; i <= attempts; i++ {
		resp, err := g.client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableTransportError(err) || i == attempts {
			break
		}

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(time.Duration(i) * 800 * time.Millisecond):
		}
	}

	return nil, lastErr
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls handshake timeout") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "temporary failure")
}

func buildPrompt(message, audience string) string {
	return fmt.Sprintf(
		"Создай рекламный баннер для аудитории \"%s\" по сообщению: \"%s\". Стиль: современный, чистый, контрастный, подходящий для веб-сайта и социальных сетей.",
		strings.TrimSpace(audience),
		strings.TrimSpace(message),
	)
}

func parseImageResult(raw map[string]any) (GenerateResult, error) {
	if imageURL := getStringByPath(raw, "data", 0, "url"); imageURL != "" {
		return GenerateResult{ImageURL: imageURL}, nil
	}
	if imageURL := getStringByPath(raw, "images", 0, "url"); imageURL != "" {
		return GenerateResult{ImageURL: imageURL}, nil
	}
	if imageURL := getStringByPath(raw, "result", "url"); imageURL != "" {
		return GenerateResult{ImageURL: imageURL}, nil
	}

	if imageBase := getStringByPath(raw, "data", 0, "b64_json"); imageBase != "" {
		return GenerateResult{ImageBase: imageBase}, nil
	}
	if imageBase := getStringByPath(raw, "images", 0, "b64_json"); imageBase != "" {
		return GenerateResult{ImageBase: imageBase}, nil
	}
	if imageBase := getStringByPath(raw, "result", "b64_json"); imageBase != "" {
		return GenerateResult{ImageBase: imageBase}, nil
	}
	if imageBase := getStringByPath(raw, "result", "image_base64"); imageBase != "" {
		return GenerateResult{ImageBase: imageBase}, nil
	}

	return GenerateResult{}, errors.New("image response does not contain url or base64")
}

func getStringByPath(value any, keys ...any) string {
	current := value
	for _, key := range keys {
		switch k := key.(type) {
		case string:
			next, ok := current.(map[string]any)
			if !ok {
				return ""
			}
			current = next[k]
		case int:
			next, ok := current.([]any)
			if !ok || k < 0 || k >= len(next) {
				return ""
			}
			current = next[k]
		default:
			return ""
		}
	}

	str, ok := current.(string)
	if !ok {
		return ""
	}
	return str
}
