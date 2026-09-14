package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStore_Create(t *testing.T) {
	var capturedMethod, capturedPath string
	var capturedBody Product

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")
	p := Product{ID: "p1", Name: "Mouse", Price: 199.9, Stock: 5}

	if err := store.Create(context.Background(), p); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if capturedMethod != http.MethodPut {
		t.Fatalf("expected PUT, got %s", capturedMethod)
	}
	if capturedPath != "/products/_doc/p1" {
		t.Fatalf("expected /products/_doc/p1, got %s", capturedPath)
	}
	if capturedBody != p {
		t.Fatalf("expected body %+v, got %+v", p, capturedBody)
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")

	_, err := store.Get(context.Background(), "missing")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_Get_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"_source": Product{ID: "p1", Name: "Monitor", Price: 999, Stock: 2},
		})
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")

	p, err := store.Get(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if p.Name != "Monitor" {
		t.Fatalf("expected Monitor, got %s", p.Name)
	}
}

func TestStore_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/_search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"hits": map[string]any{
				"hits": []map[string]any{
					{"_source": Product{ID: "p1", Name: "Kulaklık"}},
					{"_source": Product{ID: "p2", Name: "Kulaklık Standı"}},
				},
			},
		})
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")

	results, err := store.Search(context.Background(), "kulaklık")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestStore_DecrementStock(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")

	if err := store.DecrementStock(context.Background(), "p1", 3); err != nil {
		t.Fatalf("DecrementStock returned error: %v", err)
	}
	if capturedPath != "/products/_update/p1" {
		t.Fatalf("expected /products/_update/p1, got %s", capturedPath)
	}
	script, ok := capturedBody["script"].(map[string]any)
	if !ok {
		t.Fatalf("expected script in request body, got %+v", capturedBody)
	}
	params, ok := script["params"].(map[string]any)
	if !ok || params["qty"] != float64(3) {
		t.Fatalf("expected qty param 3, got %+v", script["params"])
	}
}

func TestStore_DecrementStock_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := NewStore(srv.URL, "products")

	if err := store.DecrementStock(context.Background(), "p1", 1); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
