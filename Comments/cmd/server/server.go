package main

import (
	"Comments/pkg/api"
	"Comments/pkg/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	db  *db.DB
	api *api.API
}

func main() {
	var srv server
	var err error
	srv.db, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	srv.api = api.New(srv.db)
	port := ":8082"
	logFilePath := "comments.log"

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.Printf("[*] HTTP Comments server is started on http://localhost%s", port)
	log.SetOutput(file)

	go func() {
		if err := http.ListenAndServe(port, srv.api.Router()); err != nil {
			log.Fatalf("Could not listen on port %s: %v\n", port, err)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal
	log.SetOutput(os.Stdout)
	log.Printf("[*] HTTP Comments server has been stopped. Reason: got sigterm")
}
