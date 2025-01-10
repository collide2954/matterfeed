// cmd/main.go
package main

import (
	"context"
	"database/sql"
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
	"matterfeed/messenger"
)

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
		log.Printf("Error initializing database: %v", initDBErr)
		os.Exit(1)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}(db)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	var wg sync.WaitGroup

	cfg, loadConfigErr := config.LoadConfig(configFile)
	if loadConfigErr != nil {
		log.Printf("Error reading config from file %s: %v", configFile, loadConfigErr)
		os.Exit(1)
	}

	feedHandler := feed.NewFeedHandler(feed.Config{
		URLs:        cfg.Feeds.URLs,
		RescanDelay: cfg.Feeds.RescanDelay,
	}, db)

	wg.Add(1)
	go func() {
		defer wg.Done()
		feedHandler.CheckFeeds(ctx, func(title, link string) error {
			message := fmt.Sprintf("New article: %s - %s", title, link)
			log.Println(message)
			sendMessageErr := messenger.SendMessage(cfg.Mattermost.SecretURL, message)
			if sendMessageErr != nil {
				return fmt.Errorf("Error sending message: %v", sendMessageErr)
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
		log.Println("Received shutdown signal")
		close(stopCh)
		cancel()
	}()

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	<-doneCh
	log.Println("All goroutines finished, shutting down.")
}
