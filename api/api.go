// api/api.go
package api

import (
	"context"
	"errors"
	"log"
	"matterfeed/config"
	"net/http"
	"strconv"
	"time"
)

const ReadHeaderTimeout = 5 * time.Second

type HealthCheckResponse struct {
	Status string `json:"status"`
	Port   int    `json:"port"`
}

func StartAPIServer(cfg *config.Config, stopCh <-chan struct{}) {
	port := cfg.API.Port
	log.Printf("Starting API server on port %d", port)

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error starting API server: %v", err)
		}
	}()

	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), ReadHeaderTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down API server: %v", err)
	}
}
