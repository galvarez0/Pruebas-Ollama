package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

func main() {
	model := flag.String("model", "", "modelo")
	flag.Parse()

	if *model == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	messages := []ollama.ChatMessage{
		{Role: "user", Content: "Mi nombre es Gabriel y trabajo con Go."},
		{Role: "assistant", Content: "Entendido. Te llamas Gabriel y trabajas con Go."},
		{Role: "user", Content: "¿Cómo me llamo y con qué lenguaje trabajo? Responde en una sola frase."},
	}

	resp, err := client.Chat(ctx, ollama.ChatRequest{
		Model:    *model,
		Messages: messages,
		Options: map[string]any{
			"temperature": 0,
		},
		Stream: false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR en chat: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Message.Content)
	fmt.Fprintf(os.Stderr, "metrics total=%s out_tokens=%d tok/s=%.2f\n",
		time.Duration(resp.TotalDuration),
		resp.EvalCount,
		ollama.TokensPerSecond(resp.EvalCount, resp.EvalDuration),
	)
}
