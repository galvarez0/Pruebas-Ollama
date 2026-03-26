package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprintf("%v", *m)
}

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func cosine(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func main() {
	model := flag.String("model", "", "modelo de embeddings")
	var inputs multiFlag
	flag.Var(&inputs, "input", "texto a embeber; se puede repetir varias veces")
	flag.Parse()

	if *model == "" || len(inputs) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model y al menos dos -input")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.Embed(ctx, ollama.EmbedRequest{
		Model:    *model,
		Input:    []string(inputs),
		Truncate: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR generando embeddings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Se generaron %d embeddings\n", len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		fmt.Printf("[%d] len=%d text=%q\n", i, len(emb), inputs[i])
	}

	fmt.Println("Similitudes coseno:")
	for i := 0; i < len(resp.Embeddings); i++ {
		for j := i + 1; j < len(resp.Embeddings); j++ {
			fmt.Printf("  (%d,%d) %.4f  %q <-> %q\n",
				i, j, cosine(resp.Embeddings[i], resp.Embeddings[j]), inputs[i], inputs[j])
		}
	}
}
