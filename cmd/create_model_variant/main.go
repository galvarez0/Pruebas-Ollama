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
	from := flag.String("from", "", "modelo base")
	model := flag.String("model", "", "nombre del nuevo modelo")
	system := flag.String("system", "Responde de forma breve, técnica y en español.", "system prompt")
	quantize := flag.String("quantize", "", "cuantización opcional, por ejemplo q4_K_M")
	flag.Parse()

	if *from == "" || *model == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -from y -model")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	resp, err := client.Create(ctx, ollama.CreateRequest{
		Model:  *model,
		From:   *from,
		System: *system,
		Quantize: *quantize,
		Parameters: map[string]any{
			"temperature": 0.2,
			"num_ctx":     2048,
		},
		Messages: []map[string]interface{}{
			{"role": "user", "content": "¿En qué idioma respondes?"},
			{"role": "assistant", "content": "Siempre respondo en español, salvo que me pidan otro idioma."},
		},
		Stream: false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR creando modelo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Modelo creado: %s (base=%s) status=%s\n", *model, *from, resp.Status)
}
