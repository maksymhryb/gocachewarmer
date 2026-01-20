package pipeline

import (
	"context"
	"log"

	"github.com/maksymhryb/gocachewarmer/config"
)

func Count(ctx context.Context, config *config.Config, inputCh <-chan string, refreshRate int) <-chan int {
	outputCh := make(chan int)

	go func() {
		defer log.Println("[Counter Stage] stage")
		defer close(outputCh)

		log.Println("[Counter Stage] started")

		counter := 0
		notify := func() {
			log.Println("[Counter Stage] notify")

			select {
			case outputCh <- counter:
			case <-ctx.Done():
				return
			}
		}

		for range inputCh {
			counter++
			if refreshRate != -1 && counter%refreshRate == 0 {
				notify()
			}
		}
		notify()
	}()

	return outputCh
}
