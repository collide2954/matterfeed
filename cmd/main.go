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
	"matterfeed/data"
	"matterfeed/feed"
	"matterfeed/messenger"
)

func main() {
	flag.Parse()
	configPath := flag.String("config", "", "Path to the TOML configuration file")
	cfg, loadConfigErr := config.LoadConfig(*configPath)
	if loadConfigErr != nil {
		log.Fatalf("Error loading config: %v", loadConfigErr)
	}

	db, initDBErr := data.InitDBWithRetry()
	if initDBErr != nil {
		log.Fatalf("Error initializing database: %v", initDBErr)
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
				return fmt.Errorf("error sending message: %w", sendMessageErr)
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
