// feed/feed.go
package feed

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/SlyMarbo/rss"
)

type Config struct {
	URLs        []string
	RescanDelay int
}

type Handler struct {
	config Config
	db     *sql.DB
}

func NewFeedHandler(config Config, db *sql.DB) *Handler {
	return &Handler{
		config: config,
		db:     db,
	}
}

func (fh *Handler) CheckFeeds(ctx context.Context, onNewArticle func(title, link string) error) {
	ticker := time.NewTicker(time.Duration(fh.config.RescanDelay) * time.Second)
	defer ticker.Stop()

	programStartTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("starting feed scan")
			fh.processFeeds(programStartTime, onNewArticle)
		}
	}
}

func (fh *Handler) processFeeds(programStartTime time.Time, onNewArticle func(title, link string) error) {
	for _, feedURL := range fh.config.URLs {
		feed, fetchErr := rss.Fetch(feedURL)
		if fetchErr != nil {
			log.Printf("failed fetching feed: %v", fetchErr)
			continue
		}
		fh.processFeedItems(feed.Items, programStartTime, onNewArticle)
	}
}

func (fh *Handler) processFeedItems(
	items []*rss.Item, programStartTime time.Time, onNewArticle func(title, link string) error) {
	for _, item := range items {
		if fh.isNewArticle(item, programStartTime) {
			onNewArticleErr := onNewArticle(item.Title, item.Link)
			if onNewArticleErr != nil {
				continue
			}
			fh.markArticleAsSeen(item)
		}
	}
}

func (fh *Handler) isNewArticle(item *rss.Item, programStartTime time.Time) bool {
	var seen bool
	queryErr := fh.db.QueryRow("SELECT EXISTS(SELECT 1 FROM seen_articles WHERE id = ?)", item.ID).Scan(&seen)
	if queryErr != nil {
		log.Printf("failed querying seen articles: %v", queryErr)
		return false
	}
	return !seen && item.Date.After(programStartTime)
}

func (fh *Handler) markArticleAsSeen(item *rss.Item) {
	_, insertErr := fh.db.Exec(
		"INSERT INTO seen_articles (id, title, link, date) VALUES (?, ?, ?, ?)",
		item.ID, item.Title, item.Link, item.Date)
	if insertErr != nil {
		log.Printf("failed inserting seen article: %v", insertErr)
	}
}
