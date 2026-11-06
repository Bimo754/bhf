package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMProvider represents supported AI model providers.
type LLMProvider string

const (
	ProviderMock   LLMProvider = "mock"
	ProviderGemini LLMProvider = "gemini"
	ProviderOpenAI LLMProvider = "openai"
	ProviderOllama LLMProvider = "ollama"
)

// Config specifies AI connection and model parameters.
type Config struct {
	Provider LLMProvider   `json:"provider"`
	APIKey   string        `json:"api_key"`
	Endpoint string        `json:"endpoint,omitempty"`
	Model    string        `json:"model"`
	Timeout  time.Duration `json:"timeout"`
}

// Client defines the interface for interacting with language models.
type Client interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// MockClient provides deterministic simulated responses for automated test suites.
type MockClient struct {
	ResponseHandler func(system, user string) string
}

func NewMockClient() *MockClient {
	return &MockClient{
		ResponseHandler: func(system, user string) string {
			return `{"proposals":[{"endpoint":"/api/v1/users","method":"GET","test_type":"idor","reasoning":"user_id parameter observed with integer values","priority":1}]}`
		},
	}
}

func (m *MockClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if m.ResponseHandler != nil {
		return m.ResponseHandler(systemPrompt, userPrompt), nil
	}
	return "mock response", nil
}

// HTTPClient connects to standard HTTP REST inference endpoints.
type HTTPClient struct {
	config     Config
	httpClient *http.Client
}

// NewClient constructs an appropriate Client implementation based on configuration.
func NewClient(cfg Config) (Client, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Provider == ProviderMock || (cfg.APIKey == "" && cfg.Provider != ProviderOllama) {
		return NewMockClient(), nil
	}

	return &HTTPClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// Generate dispatches an inference request to the configured provider.
func (c *HTTPClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	switch c.config.Provider {
	case ProviderOllama:
		return c.generateOllama(ctx, systemPrompt, userPrompt)
	case ProviderOpenAI, ProviderGemini:
		return c.generateOpenAICompatible(ctx, systemPrompt, userPrompt)
	default:
		return "", errors.New("unsupported AI provider: " + string(c.config.Provider))
	}
}

func (c *HTTPClient) generateOllama(ctx context.Context, system, user string) (string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "http://127.0.0.1:11434/api/generate"
	}

	reqPayload := map[string]any{
		"model":  c.config.Model,
		"system": system,
		"prompt": user,
		"stream": false,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

func (c *HTTPClient) generateOpenAICompatible(ctx context.Context, system, user string) (string, error) {
	endpoint := c.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	reqPayload := map[string]any{
		"model": c.config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var oaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return "", errors.New("empty response choices from AI provider")
	}

	return oaiResp.Choices[0].Message.Content, nil
}
