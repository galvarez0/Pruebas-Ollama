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

type StatusResponse struct {
	Status string `json:"status"`
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
