package healthcheck

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func respond(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprint(w, "ok")
}

func Start(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid health check port number: %d", port)
	}

	address := ":" + strconv.Itoa(port)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", respond)
	log.Infof("Listening for health checks on 0.0.0.0%s/healthcheck", address)
	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}
