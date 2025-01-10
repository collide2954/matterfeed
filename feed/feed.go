// feed/feed.go
package feed

import (
	"context"
	"database/sql"
	"time"

	"matterfeed/logger"

	"github.com/SlyMarbo/rss"
)

type FeedConfig struct {
	URLs        []string
	RescanDelay int
}

type FeedHandler struct {
	config FeedConfig
	db     *sql.DB
}

func NewFeedHandler(config FeedConfig, db *sql.DB) *FeedHandler {
	return &FeedHandler{
		config: config,
		db:     db,
	}
}

func (fh *FeedHandler) CheckFeeds(ctx context.Context, onNewArticle func(title, link string) error) {
	ticker := time.NewTicker(time.Duration(fh.config.RescanDelay) * time.Second)
	defer ticker.Stop()

	programStartTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.LogInfo("Starting feed scan")
			for _, feedURL := range fh.config.URLs {
				feed, fetchErr := rss.Fetch(feedURL)
				if fetchErr != nil {
					logger.LogAndReturnError(fetchErr, "fetching feed")
					continue
				}

				for _, item := range feed.Items {
					var seen bool
					queryErr := fh.db.QueryRow("SELECT EXISTS(SELECT 1 FROM seen_articles WHERE id = ?)", item.ID).Scan(&seen)
					if queryErr != nil {
						logger.LogAndReturnError(queryErr, "querying seen articles")
						continue
					}

					if !seen && item.Date.After(programStartTime) {
						onNewArticleErr := onNewArticle(item.Title, item.Link)
						if onNewArticleErr != nil {
							continue
						}

						_, insertErr := fh.db.Exec("INSERT INTO seen_articles (id, title, link, date) VALUES (?, ?, ?, ?)", item.ID, item.Title, item.Link, item.Date)
						if insertErr != nil {
							logger.LogAndReturnError(insertErr, "inserting seen article")
							continue
						}
					}
				}
			}
		}
	}
}
