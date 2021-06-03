package main

import (
	"github.com/joho/godotenv"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"gitlab.starlink.ua/high-school-prod/chat/server/seshandler"
	"net/http"
	"os"
)

// Loads values from .env into the system
func init() {
	if err := godotenv.Load(); err != nil {
		log.Logger.Fatal(err)
	}
}

//the entry point to the program which launches a session handler
func main() {

	//go tracer.RunMonitoring()

	sh := seshandler.New()

	// Handle url pattern "/chat" with "handler" function
	http.HandleFunc("/chat", sh.Handle)

	chatRoot, _ := os.LookupEnv("SOCKET")

	log.Logger.Infof("Listening and serving on %s", chatRoot)
	// Log an error if the server failed to start
	log.Logger.Fatal(http.ListenAndServe(chatRoot, nil))
}
