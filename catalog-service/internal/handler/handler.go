package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/example/cloud-microservices-showcase/catalog-service/internal/catalog"
	"github.com/google/uuid"
)

type Handler struct {
	store *catalog.Store
}

func New(store *catalog.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var p catalog.Product
	if err := json.Unmarshal(body, &p); err != nil || p.Name == "" {
		http.Error(w, "invalid product: 'name' is required", http.StatusBadRequest)
		return
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	}

	if err := h.store.Create(r.Context(), p); err != nil {
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := h.store.Get(r.Context(), id)
	if errors.Is(err, catalog.ErrNotFound) {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch product", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := h.store.Search(r.Context(), q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
