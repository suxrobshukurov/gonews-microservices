package main

import (
	"APIGateway/pkg/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	api *api.API
}

func main() {
	var srv server
	srv.api = api.New()

	port := ":8080"
	logFilePath := "apigateway.log"
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.Printf("[*] HTTP APIGateway server is started on http://localhost%s", port)
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
	log.Printf("[*] HTTP APIGateway server has been stopped. Reason: got sigterm")
}
