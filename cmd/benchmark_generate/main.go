package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/galvarez0/Pruebas-Ollama/internal/ollama"
)

type sample struct {
	latency time.Duration
	tokens  int64
	tps     float64
	err     error
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func main() {
	model := flag.String("model", "", "modelo")
	prompt := flag.String("prompt", "Explica qué es un benchmark en una frase.", "prompt")
	n := flag.Int("n", 10, "número total de requests")
	c := flag.Int("c", 2, "concurrencia")
	flag.Parse()

	if *model == "" {
		fmt.Fprintln(os.Stderr, "ERROR: debes pasar -model")
		os.Exit(2)
	}
	if *n <= 0 || *c <= 0 {
		fmt.Fprintln(os.Stderr, "ERROR: -n y -c deben ser > 0")
		os.Exit(2)
	}

	client := ollama.NewFromEnv()
	jobs := make(chan int)
	results := make(chan sample, *n)

	var started int64
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for range jobs {
			reqNum := atomic.AddInt64(&started, 1)
			_ = reqNum

			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			t0 := time.Now()
			resp, err := client.Generate(ctx, ollama.GenerateRequest{
				Model:  *model,
				Prompt: *prompt,
				Options: map[string]any{
					"temperature": 0,
				},
				Stream: false,
			})
			cancel()

			if err != nil {
				results <- sample{err: err}
				continue
			}

			results <- sample{
				latency: time.Since(t0),
				tokens:  resp.EvalCount,
				tps:     ollama.TokensPerSecond(resp.EvalCount, resp.EvalDuration),
			}
		}
	}

	for i := 0; i < *c; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for i := 0; i < *n; i++ {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var (
		latencies []time.Duration
		totalTok  int64
		totalTPS  float64
		okCount   int
		errCount  int
	)

	for r := range results {
		if r.err != nil {
			errCount++
			fmt.Fprintf(os.Stderr, "request error: %v\n", r.err)
			continue
		}
		okCount++
		totalTok += r.tokens
		totalTPS += r.tps
		latencies = append(latencies, r.latency)
	}

	if okCount == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: todas las requests fallaron")
		os.Exit(1)
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}

	fmt.Printf("requests_ok=%d requests_error=%d\n", okCount, errCount)
	fmt.Printf("latency_avg=%s latency_p50=%s latency_p95=%s latency_max=%s\n",
		sum/time.Duration(okCount),
		percentile(latencies, 0.50),
		percentile(latencies, 0.95),
		latencies[len(latencies)-1],
	)
	fmt.Printf("tokens_total=%d tps_avg=%.2f\n", totalTok, totalTPS/float64(okCount))
}
