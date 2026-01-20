package pipeline

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/maksymhryb/gocachewarmer/config"
)

func FilterInput(ctx context.Context, config *config.Config, inputCh <-chan string) <-chan string {
	outputCh := make(chan string)
	go func() {
		defer close(outputCh)
		defer log.Println("[Filter Input Stage] stopped")
		log.Println("[Filter Input Stage] started")

		var re *regexp.Regexp
		if config.FilterInputUrl != "" {
			parts := strings.Split(config.FilterInputUrl, "*")
			for i, v := range parts {
				parts[i] = regexp.QuoteMeta(v)
			}
			re = regexp.MustCompile(strings.Join(parts, ".*"))
		}
	loop:
		for url := range inputCh {
			if re != nil && !re.MatchString(url) {
				continue loop
			}
			select {
			case outputCh <- url:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputCh
}
