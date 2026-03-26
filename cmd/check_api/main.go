package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

func main() {
	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := client.Tags(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: no se pudo conectar a %s: %v\n", client.BaseURL(), err)
		os.Exit(1)
	}

	fmt.Printf("OK: API disponible en %s\n", client.BaseURL())
	fmt.Printf("Modelos instalados: %d\n", len(resp.Models))
}
