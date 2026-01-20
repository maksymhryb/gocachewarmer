package pipeline

import (
	"context"
	"log"
	"sync"

	"github.com/maksymhryb/gocachewarmer/config"
)

func TeeN[T any](ctx context.Context, config *config.Config, inputCh <-chan T, n int, buffSize int) []<-chan T {
	outs := make([]chan T, n)
	for i := range outs {
		outs[i] = make(chan T, buffSize)
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
			log.Println("[TeeN Stage] stopped")
		}()

		log.Println("[TeeN Stage] started")

		for val := range inputCh {
			log.Println("[TeeN Stage] duplicating ")

			var wg sync.WaitGroup
			wg.Add(n)

			for i := range outs {
				go func(ch chan T, v T) {
					defer wg.Done()
					ch <- v
				}(outs[i], val)
			}

			wg.Wait()
		}
	}()

	result := make([]<-chan T, n)
	for i, ch := range outs {
		result[i] = ch
	}

	return result
}
