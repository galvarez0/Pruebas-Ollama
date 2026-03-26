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
	prompt := flag.String("prompt", "", "prompt")
	out := flag.String("out", "output.txt", "archivo destino")
	flag.Parse()

	if *model == "" || *prompt == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model y -prompt")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := client.Generate(ctx, ollama.GenerateRequest{
		Model:  *model,
		Prompt: *prompt,
		Stream: false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generando texto: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(resp.Response), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR escribiendo archivo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Respuesta guardada en %s\n", *out)
}
