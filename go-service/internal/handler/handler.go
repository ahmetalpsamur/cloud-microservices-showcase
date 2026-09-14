package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/example/cloud-microservices-showcase/go-service/internal/rabbitmq"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Handler struct {
	publisher *rabbitmq.Publisher
}

func New(publisher *rabbitmq.Publisher) *Handler {
	return &Handler{publisher: publisher}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var evt Event
	if err := json.Unmarshal(body, &evt); err != nil || evt.Type == "" {
		http.Error(w, "invalid event: 'type' is required", http.StatusBadRequest)
		return
	}

	if err := h.publisher.Publish(r.Context(), evt.Type, body); err != nil {
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}
