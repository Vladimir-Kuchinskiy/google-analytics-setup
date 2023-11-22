package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	runGracefulShutDownListener := func() {
		osCall := <-c
		logger.Info("stop syscall", "code", osCall.String())
		ticker.Stop()
		cancel()
	}

	go runGracefulShutDownListener()

	// Initialize rates API client
	ratesAPIClient := NewRatesAPIClient()

	apiToken := "FKS82y8KTCCzGxgnJc3tqQ"
	measurementID := "G-112ZC4DE4F"

	googleAnalyticsAPIClient := NewGoogleAnalyticsAPIClient(apiToken)

	pushEvent := func(timestamp time.Time) error {
		ratio, err := ratesAPIClient.fetchUAHtoUSDCurrenciesRatio()
		if err != nil {
			return err
		}

		event := Event{
			Name: "uah_to_usd",
			Params: map[string]interface{}{
				"rate":                 ratio.Rate,
				"session_id":           "123",
				"engagement_time_msec": "100",
				"timestamp":            timestamp.Format(time.DateTime),
			},
		}

		logger.Info("pushed event", "event", event)

		if err = googleAnalyticsAPIClient.pushEvent(measurementID, "785147521.1700385219", []Event{event}); err != nil {
			return err
		}

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if err := pushEvent(t.UTC()); err != nil {
				panic(err)
			}
		}
	}
}
