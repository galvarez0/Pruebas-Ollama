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

	resp, err := client.PS(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR listando modelos en memoria: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Models) == 0 {
		fmt.Println("No hay modelos cargados en memoria.")
		return
	}

	for _, m := range resp.Models {
		fmt.Printf(
			"%-35s ctx=%-6d vram=%-10s family=%-12s expires=%s\n",
			m.Name,
			m.ContextLength,
			ollama.HumanBytes(m.SizeVRAM),
			m.Details.Family,
			m.ExpiresAt,
		)
	}
}
