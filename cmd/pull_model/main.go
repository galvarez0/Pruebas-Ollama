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
	model := flag.String("model", "", "modelo a descargar")
	insecure := flag.Bool("insecure", false, "permitir descarga insegura")
	flag.Parse()

	if *model == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	resp, err := client.Pull(ctx, ollama.PullRequest{
		Model:    *model,
		Insecure: *insecure,
		Stream:   false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR descargando modelo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pull completado para %s: %s\n", *model, resp.Status)
}
