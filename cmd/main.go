// cmd/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"matterfeed/api"
	"matterfeed/config"
	"matterfeed/feed"
	"matterfeed/logger"
	"matterfeed/messenger"
)

type Message struct {
	Text string `json:"text"`
}

var (
	ConfigFile = flag.String("config", "", "Valid TOML configuration file")
)

func main() {
	flag.Parse()

	configFile, getConfigErr := config.GetSingleConfigFile(*ConfigFile)
	if getConfigErr != nil {
		log.Fatalf("Error getting config file: %v", getConfigErr)
	}

	db, initDBErr := InitDBWithRetry()
	if initDBErr != nil {
		logger.LogAndReturnError(initDBErr, "initializing database")
		os.Exit(1)
	}
	defer db.Close()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	var wg sync.WaitGroup

	cfg, loadConfigErr := config.LoadConfig(configFile)
	if loadConfigErr != nil {
		logger.LogAndReturnError(loadConfigErr, fmt.Sprintf("reading config from file: %s", configFile))
		os.Exit(1)
	}

	logger.InitLogger(cfg.Logging.OutputToTerminal)

	feedHandler := feed.NewFeedHandler(feed.FeedConfig{
		URLs:        cfg.Feeds.URLs,
		RescanDelay: cfg.Feeds.RescanDelay,
	}, db)

	wg.Add(1)
	go func() {
		defer wg.Done()
		feedHandler.CheckFeeds(ctx, func(title, link string) error {
			message := fmt.Sprintf("New article: %s - %s", title, link)
			logger.LogInfo(message)
			sendMessageErr := messenger.SendMessage(cfg.Mattermost.SecretURL, message)
			if sendMessageErr != nil {
				return logger.LogAndReturnError(sendMessageErr, "sending message")
			}
			return nil
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		api.StartAPIServer(cfg, stopCh)
	}()

	go func() {
		<-signalChan
		fmt.Println("Received shutdown signal")
		close(stopCh)
		cancel()
	}()

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	<-doneCh
	fmt.Println("All goroutines finished, shutting down.")
}
