package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

func main() {
	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Tags(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR listando modelos: %v\n", err)
		os.Exit(1)
	}

	sort.Slice(resp.Models, func(i, j int) bool {
		return resp.Models[i].Name < resp.Models[j].Name
	})

	if len(resp.Models) == 0 {
		fmt.Println("No hay modelos instalados.")
		return
	}

	for _, m := range resp.Models {
		fmt.Printf(
			"%-35s size=%-10s family=%-12s params=%-8s quant=%-8s modified=%s\n",
			m.Name,
			ollama.HumanBytes(m.Size),
			m.Details.Family,
			m.Details.ParameterSize,
			m.Details.QuantizationLevel,
			m.ModifiedAt,
		)
	}
}
