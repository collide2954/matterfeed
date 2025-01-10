// api/api.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"matterfeed/config"
	"matterfeed/logger"
)

type HealthCheckResponse struct {
	Status string `json:"status"`
	Port   int    `json:"port"`
}

func StartAPIServer(cfg *config.Config, stopCh <-chan struct{}) {
	port := cfg.API.Port
	logger.LogInfo(fmt.Sprintf("Starting API server on port %d", port))

	srv := &http.Server{
		Addr: ":" + strconv.Itoa(port),
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(err, "starting API server")
		}
	}()

	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.LogError(err, "shutting down API server")
	}
}
