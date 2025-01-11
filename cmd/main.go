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
	configPath := flag.String("config", "", "Path to the TOML configuration file")
	flag.Parse()
	cfg, loadConfigErr := config.LoadConfig(*configPath)
	if loadConfigErr != nil {
		log.Fatalf("error loading config: %v", loadConfigErr)
	}

	db, initDBErr := data.InitDBWithRetry()
	if initDBErr != nil {
		log.Fatalf("error initializing database: %v", initDBErr)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("error closing database: %v", err)
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
			message := fmt.Sprintf("New Article: %s - %s", title, link)
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
		log.Println("Received Shutdown Signal")
		close(stopCh)
		cancel()
	}()

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	<-doneCh
	log.Println("All Goroutines Finished, Shutting Down.")
}
