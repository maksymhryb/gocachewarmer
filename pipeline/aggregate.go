package pipeline

import (
	"context"
	"log"
	"time"

	"github.com/maksymhryb/gocachewarmer/config"
)

type AggregateResult struct {
	AvgResponseTime   time.Duration
	StatusCodeCount   map[int]int
	TotalResponses    int
	SuccessResponses  int
	ErrorResponses    int
	RedirectResponses int
	SuccessRate       float64
}

func Aggregate(ctx context.Context, config *config.Config, inputCh <-chan WarmupResult) <-chan AggregateResult {
	outputCh := make(chan AggregateResult)

	go func() {
		defer close(outputCh)
		defer log.Println("[Aggregate Stage] stopped")

		log.Println("[Aggregate Stage] started")

		prevAgregateResult := &AggregateResult{
			StatusCodeCount: make(map[int]int),
		}
		var aggregateResult AggregateResult
		for val := range inputCh {
			log.Printf("[Aggregate Stage] processing url \"%s\"\n", val.Url)

			aggregateResult = *prevAgregateResult

			if aggregateResult.AvgResponseTime > 0 {
				aggregateResult.AvgResponseTime = aggregateResult.AvgResponseTime + (val.ResponseTime-aggregateResult.AvgResponseTime)/time.Duration(aggregateResult.TotalResponses)
			} else {
				aggregateResult.AvgResponseTime = val.ResponseTime
			}

			if _, exist := aggregateResult.StatusCodeCount[val.StatusCode]; !exist {
				aggregateResult.StatusCodeCount[val.StatusCode] = 0
			}
			aggregateResult.StatusCodeCount[val.StatusCode]++

			aggregateResult.TotalResponses++
			if val.StatusCode >= 200 && val.StatusCode < 300 {
				aggregateResult.SuccessResponses++
			} else if val.StatusCode >= 300 && val.StatusCode < 400 {
				aggregateResult.RedirectResponses++
			} else if val.StatusCode >= 500 && val.StatusCode < 600 {
				aggregateResult.ErrorResponses++
			}
			aggregateResult.SuccessRate = float64(aggregateResult.TotalResponses-aggregateResult.ErrorResponses) / float64(aggregateResult.TotalResponses)

			select {
			case outputCh <- aggregateResult:
				prevAgregateResult = &aggregateResult
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputCh
}
