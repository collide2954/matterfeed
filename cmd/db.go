// cmd/db.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	connectionString := "file:matterfeed.db?" +
		"_pragma=journal_mode(WAL)&" +
		"_pragma=busy_timeout(5000)&" +
		"_pragma=synchronous(NORMAL)&" +
		"_pragma=cache_size(2000)&" +
		"_pragma=temp_store(memory)&" +
		"_pragma=foreign_keys(true)&" +
		"_pragma=analysis_limit(400)"

	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS seen_articles (
        id TEXT PRIMARY KEY,
        title TEXT,
        link TEXT,
        date DATETIME
    )`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitDBWithRetry() (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = InitDB()
		if err == nil {
			return db, nil
		}
		if errors.Is(err, sql.ErrConnDone) || err.Error() == "database is locked" {
			fmt.Printf("Database is locked, retrying... (%d/5)\n", i+1)
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("failed to initialize database after retries: %w", err)
}