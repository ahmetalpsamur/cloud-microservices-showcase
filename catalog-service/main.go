package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/cloud-microservices-showcase/catalog-service/internal/catalog"
	"github.com/example/cloud-microservices-showcase/catalog-service/internal/handler"
	"github.com/example/cloud-microservices-showcase/catalog-service/internal/rabbitmq"
)

func main() {
	amqpURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	esURL := getenv("ELASTICSEARCH_URL", "http://localhost:9200")
	addr := getenv("HTTP_ADDR", ":8080")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := catalog.NewStore(esURL, "products")
	if err := store.EnsureIndex(ctx); err != nil {
		log.Fatalf("failed to ensure elasticsearch index: %v", err)
	}

	consumer, err := rabbitmq.NewConsumer(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer consumer.Close()

	go func() {
		if err := consumer.Run(ctx, store.DecrementStock); err != nil {
			log.Printf("stock-sync consumer stopped: %v", err)
		}
	}()

	h := handler.New(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("POST /products", h.CreateProduct)
	mux.HandleFunc("GET /products/search", h.SearchProducts)
	mux.HandleFunc("GET /products/{id}", h.GetProduct)

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	log.Printf("catalog-service listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
