package healthcheck

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func StartHealthCheck(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("Invalid health check port number: %v", port)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthcheck)
	p := ":" + strconv.Itoa(port)
	log.Infof("Listening for health checks on 0.0.0.0%v/healthcheck", p)
	server := &http.Server{
		Addr:              p,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}
