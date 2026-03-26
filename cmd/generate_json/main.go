package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

func main() {
	model := flag.String("model", "", "modelo")
	prompt := flag.String("prompt", "", "texto a procesar")
	flag.Parse()

	if *model == "" || *prompt == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model y -prompt")
		os.Exit(2)
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": []string{"string", "null"}},
			"email": map[string]any{"type": []string{"string", "null"}},
			"phone": map[string]any{"type": []string{"string", "null"}},
		},
		"required": []string{"name", "email", "phone"},
		"additionalProperties": false,
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.Generate(ctx, ollama.GenerateRequest{
		Model:  *model,
		Prompt: *prompt,
		Format: schema,
		Options: map[string]any{
			"temperature": 0,
		},
		Stream: false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generando JSON: %v\n", err)
		os.Exit(1)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(resp.Response), &decoded); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: la respuesta no es JSON válido: %v\nRespuesta: %s\n", err, resp.Response)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(decoded, "", "  ")
	fmt.Println(string(pretty))
	fmt.Fprintf(os.Stderr, "metrics total=%s out_tokens=%d tok/s=%.2f\n",
		time.Duration(resp.TotalDuration),
		resp.EvalCount,
		ollama.TokensPerSecond(resp.EvalCount, resp.EvalDuration),
	)
}
