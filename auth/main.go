package main

import (
	"auth/database"
	"auth/helpers"
	"auth/router"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/valyala/fasthttp"
)

const (
	port = ":5000"
)

func main() {
	// Initializing of Database Connection
	database.Connect()
	defer database.Disconnect()

	syscallChan := make(chan os.Signal, 1)
	signal.Notify(syscallChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-syscallChan // goroutine will be frozed at here cause it will be wating until signal is received.
		log.Println("Shutting down...")
		database.Disconnect()
		os.Exit(0)
	}()

	// Initializing of Router and starting of Server
	router := router.NewRouter()
	log.Println("Starting server on port :", port)
	log.Fatal(fasthttp.ListenAndServe(port, helpers.CommonHandlerMiddleware(router.Handler)))
}
