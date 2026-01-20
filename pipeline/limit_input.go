package pipeline

import (
	"context"
	"log"
	"math"

	"github.com/maksymhryb/gocachewarmer/config"
)

func LimitInput(ctx context.Context, cancel context.CancelFunc, config *config.Config, inputCh <-chan string) <-chan string {
	outputCh := make(chan string)

	go func() {
		defer log.Println("[Limiter Stage] stopped")
		defer close(outputCh)

		log.Println("[Limiter Stage] started")
		limit := config.Limit
		if limit < 1 {
			limit = math.MaxInt
		}
		i := 0
		for val := range inputCh {
			select {
			case outputCh <- val:
			case <-ctx.Done():
				return
			}
			i++
			if i >= limit {
				cancel()
				return
			}
		}
	}()

	return outputCh
}
