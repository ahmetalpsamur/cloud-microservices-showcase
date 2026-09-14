package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/cloud-microservices-showcase/go-service/internal/handler"
	"github.com/example/cloud-microservices-showcase/go-service/internal/rabbitmq"
)

func main() {
	amqpURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	addr := getenv("HTTP_ADDR", ":8080")

	publisher, err := rabbitmq.NewPublisher(amqpURL, "events.exchange")
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer publisher.Close()

	h := handler.New(publisher)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Health)
	mux.HandleFunc("/events", h.CreateEvent)

	log.Printf("go-service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
