package pipeline

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/maksymhryb/gocachewarmer/config"
)

func SaveReport(ctx context.Context, config *config.Config, inputCh <-chan WarmupResult) chan struct{} {
	outputCh := make(chan struct{})

	go func() {
		defer close(outputCh)
		defer log.Println("[Save Report Stage] stopped")
		log.Println("[Save Report Stage] started")

		var writer *csv.Writer
		if config.SkipReport {
			writer = csv.NewWriter(io.Discard)
		} else {
			reportFile, err := os.Create(config.ReportName)
			if err != nil {
				panic("Cannot create report file")
			}
			defer reportFile.Close()
			writer = csv.NewWriter(reportFile)
			defer writer.Flush()
			writer.Write([]string{"#", "Url", "Status Code", "Response Time", "Error"})
		}

		i := 1
		for {
			select {
			case r, ok := <-inputCh:
				if !ok {
					return
				}
				writer.Write([]string{
					strconv.Itoa(i),
					r.Url,
					strconv.Itoa(r.StatusCode),
					r.ResponseTime.String(),
					fmt.Sprint(r.Error),
				})
				i++
				log.Printf("[Save Report Stage] processing url \"%s\"\n", r.Url)
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputCh
}

func SaveAggregatedReport(ctx context.Context, config *config.Config, inputCh <-chan AggregateResult) chan struct{} {
	outputCh := make(chan struct{})

	go func() {
		defer close(outputCh)
		defer log.Println("[Save Aggregate Report Stage] stopped")
		log.Println("[Save Aggregate Report Stage] started")

		var writer *csv.Writer
		if config.SkipAggregateReport {
			writer = csv.NewWriter(io.Discard)
		} else {
			reportFile, err := os.Create(config.AggregateReportName)
			if err != nil {
				panic("Cannot create aggregate report file")
			}
			defer reportFile.Close()
			writer = csv.NewWriter(reportFile)
			defer writer.Flush()
			writer.Write([]string{"URLs Processed", "Success URLs Count", "Redirect URLs Count", "Error URLs Count", "Success Ratio", "Avg Response Time"})
		}

		var prevResult AggregateResult
		for {
			select {
			case r, ok := <-inputCh:
				if !ok {
					writer.Write([]string{
						strconv.Itoa(prevResult.TotalResponses),
						strconv.Itoa(prevResult.SuccessResponses),
						strconv.Itoa(prevResult.RedirectResponses),
						strconv.Itoa(prevResult.ErrorResponses),
						strconv.FormatFloat(prevResult.SuccessRate, 'f', 2, 64),
						prevResult.AvgResponseTime.String(),
					})
					return
				}
				prevResult = r
			case <-ctx.Done():
				return
			}
		}

	}()

	return outputCh
}
