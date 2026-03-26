package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

func main() {
	model := flag.String("model", "", "nombre del modelo")
	verbose := flag.Bool("verbose", false, "incluir campos más grandes")
	flag.Parse()

	if *model == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Show(ctx, ollama.ShowRequest{
		Model:   *model,
		Verbose: *verbose,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR mostrando modelo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Modelo: %s\n", *model)
	fmt.Printf("Familia: %s\n", resp.Details.Family)
	fmt.Printf("Parámetros: %s\n", resp.Details.ParameterSize)
	fmt.Printf("Cuantización: %s\n", resp.Details.QuantizationLevel)
	fmt.Printf("Capacidades: %v\n", resp.Capabilities)
	fmt.Printf("Modificado: %s\n", resp.ModifiedAt)
	fmt.Printf("Parameters:\n%s\n", resp.Parameters)

	if len(resp.ModelInfo) > 0 {
		fmt.Println("ModelInfo:")
		keys := make([]string, 0, len(resp.ModelInfo))
		for k := range resp.ModelInfo {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s = %v\n", k, resp.ModelInfo[k])
		}
	}
}
