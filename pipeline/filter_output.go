package pipeline

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/maksymhryb/gocachewarmer/config"
)

func FilterOutput(ctx context.Context, config *config.Config, inputCh <-chan WarmupResult) <-chan WarmupResult {
	outputCh := make(chan WarmupResult)
	go func() {
		defer log.Println("[Filter Output Stage] stopped")
		defer close(outputCh)
		log.Println("[Filter Output Stage] started")

	loop:
		for val := range inputCh {
			if !matchOutputFilters(val, config.FilterOutputStatus, config.FilterOutputResponseTime) {
				continue loop
			}
			select {
			case outputCh <- val:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputCh
}

func matchOutputFilters(result WarmupResult, statusMask string, responseTimeThreshold time.Duration) bool {
	if result, _ := matchStatus(statusMask, result.StatusCode); !result {
		return false
	}

	if responseTimeThreshold > 0 && result.ResponseTime < responseTimeThreshold {
		return false
	}

	return true
}

func matchStatus(statusMask string, status int) (bool, error) {
	if statusMask == "" {
		return true, nil
	}
	if len(statusMask) > 3 {
		return false, errors.New("invalid status mask length")
	}

	statusString := strconv.Itoa(status)

loop:
	for i, digit := range statusMask {
		if strings.EqualFold(string(digit), "X") {
			continue loop
		}
		if !unicode.IsDigit(digit) {
			return false, errors.New("invalid status mask format")
		}
		if byte(digit) != statusString[i] {
			return false, nil
		}
	}

	return true, nil
}
