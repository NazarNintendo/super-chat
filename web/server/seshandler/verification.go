package seshandler

import (
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"net/http"
	"strings"
)

// Declaring a slice of accepted origins for WebSocket to connect from.
var acceptedOrigins = []string{
	"127.0.0.1",
	"file://",
	"localhost",
	"http://stage.hsprod.tech",
	"https://stage.hsprod.tech",
}

// checkOrigin - if accepted origins contain request origin, then allow to establish the connection.
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	for _, acceptedOrigin := range acceptedOrigins {
		if strings.Contains(origin, acceptedOrigin) {
			log.Logger.Infof("Established connection with %v", origin)
			return true
		}
	}
	return false
}
