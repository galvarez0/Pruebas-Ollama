package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewFromEnv() *Client {
	baseURL := strings.TrimRight(envOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"), "/")
	timeout := 120 * time.Second

	if raw := strings.TrimSpace(os.Getenv("OLLAMA_TIMEOUT")); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			timeout = d
		}
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) endpoint(path string) string {
	if strings.HasPrefix(path, "/") {
		return c.baseURL + path
	}
	return c.baseURL + "/" + path
}

type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ollama api error: status=%s code=%d body=%s", e.Status, e.StatusCode, e.Body)
}

func (c *Client) doJSON(ctx context.Context, method, path string, reqBody any, out any) error {
	var bodyReader io.Reader

	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(rawResp),
		}
	}

	if out == nil {
		return nil
	}
	if len(rawResp) == 0 {
		return errors.New("empty response body")
	}

	if err := json.Unmarshal(rawResp, out); err != nil {
		return fmt.Errorf("decode response: %w; body=%s", err, string(rawResp))
	}
	return nil
}

type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
	ParentModel       string   `json:"parent_model"`
}

type ModelSummary struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt string       `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

type TagsResponse struct {
	Models []ModelSummary `json:"models"`
}

func (c *Client) Tags(ctx context.Context) (*TagsResponse, error) {
	var out TagsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/tags", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type RunningModel struct {
	Name          string       `json:"name"`
	Model         string       `json:"model"`
	Size          int64        `json:"size"`
	Digest        string       `json:"digest"`
	Details       ModelDetails `json:"details"`
	ExpiresAt     string       `json:"expires_at"`
	SizeVRAM      int64        `json:"size_vram"`
	ContextLength int          `json:"context_length"`
}

type PSResponse struct {
	Models []RunningModel `json:"models"`
}

func (c *Client) PS(ctx context.Context) (*PSResponse, error) {
	var out PSResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/ps", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ShowRequest struct {
	Model   string `json:"model"`
	Verbose bool   `json:"verbose,omitempty"`
}

type ShowResponse struct {
	Parameters   string                 `json:"parameters"`
	License      string                 `json:"license"`
	Template     string                 `json:"template"`
	System       string                 `json:"system"`
	ModifiedAt   string                 `json:"modified_at"`
	Capabilities []string               `json:"capabilities"`
	Details      ModelDetails           `json:"details"`
	ModelInfo    map[string]interface{} `json:"model_info"`
}

func (c *Client) Show(ctx context.Context, req ShowRequest) (*ShowResponse, error) {
	var out ShowResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/show", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type PullRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func (c *Client) Pull(ctx context.Context, req PullRequest) (*StatusResponse, error) {
	var out StatusResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/pull", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type CreateRequest struct {
	Model      string                   `json:"model"`
	From       string                   `json:"from"`
	System     string                   `json:"system,omitempty"`
	Template   string                   `json:"template,omitempty"`
	License    any                      `json:"license,omitempty"`
	Parameters map[string]any           `json:"parameters,omitempty"`
	Messages   []map[string]interface{} `json:"messages,omitempty"`
	Quantize   string                   `json:"quantize,omitempty"`
	Stream     bool                     `json:"stream"`
}

func (c *Client) Create(ctx context.Context, req CreateRequest) (*StatusResponse, error) {
	var out StatusResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/create", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type GenerateRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	System    string         `json:"system,omitempty"`
	Suffix    string         `json:"suffix,omitempty"`
	Format    any            `json:"format,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
	Stream    bool           `json:"stream"`
	KeepAlive string         `json:"keep_alive,omitempty"`
	Raw       bool           `json:"raw,omitempty"`
}

type GenerateResponse struct {
	Model              string      `json:"model"`
	CreatedAt          string      `json:"created_at"`
	Response           string      `json:"response"`
	Thinking           string      `json:"thinking"`
	Done               bool        `json:"done"`
	DoneReason         string      `json:"done_reason"`
	TotalDuration      int64       `json:"total_duration"`
	LoadDuration       int64       `json:"load_duration"`
	PromptEvalCount    int64       `json:"prompt_eval_count"`
	PromptEvalDuration int64       `json:"prompt_eval_duration"`
	EvalCount          int64       `json:"eval_count"`
	EvalDuration       int64       `json:"eval_duration"`
	Logprobs           interface{} `json:"logprobs"`
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	var out GenerateResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/generate", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Tools       []any          `json:"tools,omitempty"`
	Format      any            `json:"format,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
	Stream      bool           `json:"stream"`
	Think       any            `json:"think,omitempty"`
	KeepAlive   string         `json:"keep_alive,omitempty"`
	Logprobs    bool           `json:"logprobs,omitempty"`
	TopLogprobs int            `json:"top_logprobs,omitempty"`
}

type ChatResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Message            struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Thinking  string `json:"thinking"`
		ToolCalls []any  `json:"tool_calls"`
	} `json:"message"`
	Done               bool  `json:"done"`
	DoneReason         string `json:"done_reason"`
	TotalDuration      int64 `json:"total_duration"`
	LoadDuration       int64 `json:"load_duration"`
	PromptEvalCount    int64 `json:"prompt_eval_count"`
	PromptEvalDuration int64 `json:"prompt_eval_duration"`
	EvalCount          int64 `json:"eval_count"`
	EvalDuration       int64 `json:"eval_duration"`
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var out ChatResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/chat", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type EmbedRequest struct {
	Model      string         `json:"model"`
	Input      any            `json:"input"`
	Truncate   bool           `json:"truncate,omitempty"`
	Dimensions *int           `json:"dimensions,omitempty"`
	KeepAlive  string         `json:"keep_alive,omitempty"`
	Options    map[string]any `json:"options,omitempty"`
}

type EmbedResponse struct {
	Model           string        `json:"model"`
	Embeddings      [][]float64   `json:"embeddings"`
	TotalDuration   int64         `json:"total_duration"`
	LoadDuration    int64         `json:"load_duration"`
	PromptEvalCount int64         `json:"prompt_eval_count"`
}

func (c *Client) Embed(ctx context.Context, req EmbedRequest) (*EmbedResponse, error) {
	var out EmbedResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/embed", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func HumanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func TokensPerSecond(evalCount, evalDurationNS int64) float64 {
	if evalCount <= 0 || evalDurationNS <= 0 {
		return 0
	}
	return float64(evalCount) / (float64(evalDurationNS) / 1e9)
}
