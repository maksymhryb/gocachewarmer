package pipeline

import (
	"context"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/maksymhryb/gocachewarmer/client"
	"github.com/maksymhryb/gocachewarmer/config"
)

const (
	MIN_DURATION = 30
	MAX_DURATION = 700
)

const (
	REDIRECT_STATUS_PROBABILITY = 0.05
	ERROR_STATUS_PROBABILITY    = 0.05
)

type WarmupResult struct {
	Url          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

func Warmup(ctx context.Context, config *config.Config, inputCh <-chan string) <-chan WarmupResult {
	outputCh := make(chan WarmupResult)

	go func() {
		defer log.Println("[Warmup Stage] stopped")
		defer close(outputCh)
		log.Println("[Warmup Stage] started")

		cl := client.NewClient(config.UserAgent, config.ConnectionTimeout)
		startWorker := func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			defer log.Printf("[Warmup Stage] worker %d stopped", i)
			log.Printf("[Warmup Stage] worker %d started", i)

			for {
				select {
				case url, ok := <-inputCh:
					if !ok {
						return
					}
					log.Printf("[Warmup Stage] processing url \"%s\"", url)
					timeStarted := time.Now()
					var result WarmupResult
					if config.DryRun {
						time.Sleep(generateRandomTimeout())
						result = WarmupResult{
							Url:          url,
							StatusCode:   generateRandomStatus(),
							ResponseTime: time.Since(timeStarted),
						}
					} else {
						response, err := cl.Get(url)
						tookTime := time.Since(timeStarted)
						io.Copy(io.Discard, response.Body)
						response.Body.Close()
						result = WarmupResult{
							Url:          url,
							StatusCode:   response.StatusCode,
							ResponseTime: tookTime,
							Error:        err,
						}
					}

					outputCh <- result
				case <-ctx.Done():
					return
				}
			}
		}

		var wg sync.WaitGroup
		for i := range config.ConcurrentRequests {
			wg.Add(1)
			startWorker(&wg, i)
		}
		wg.Wait()
	}()

	return outputCh
}

func generateRandomTimeout() time.Duration {
	mean := (MIN_DURATION + MAX_DURATION) / 2.0
	stdDev := (MAX_DURATION - MIN_DURATION) / 6.0

	value := rand.NormFloat64()*stdDev + mean
	value = math.Max(value, MIN_DURATION)
	value = math.Min(value, MAX_DURATION)

	return time.Duration(value) * time.Millisecond
}

func generateRandomStatus() int {
	r := rand.Float64()
	switch {
	case r < REDIRECT_STATUS_PROBABILITY:
		return 301
	case r < (REDIRECT_STATUS_PROBABILITY + ERROR_STATUS_PROBABILITY):
		return 500
	default:
		return 200
	}
}
