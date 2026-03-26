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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tests := []string{
		"Responde solo: ok",
		"Cuenta del 1 al 3",
		"Devuelve un JSON válido: {\"status\":\"ok\"}",
	}

	for i, p := range tests {
		resp, err := client.Generate(ctx, ollama.GenerateRequest{
			Model:  *model,
			Prompt: p,
			Options: map[string]any{
				"temperature": 0,
			},
			Stream: false,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL test=%d err=%v\n", i+1, err)
			os.Exit(1)
		}
		fmt.Printf("TEST %d\nPROMPT: %s\nRESPUESTA: %s\n---\n", i+1, p, resp.Response)
	}

	fmt.Println("Smoke suite completada.")
}
