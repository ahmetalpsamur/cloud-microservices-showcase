package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/order-catalog-microservices/catalog-service/internal/catalog"
)

// fakeStore is an in-memory ProductStore used to unit test the HTTP layer
// without talking to Elasticsearch.
type fakeStore struct {
	products map[string]catalog.Product
	searchFn func(query string) []catalog.Product
}

func newFakeStore() *fakeStore {
	return &fakeStore{products: map[string]catalog.Product{}}
}

func (f *fakeStore) Create(_ context.Context, p catalog.Product) error {
	f.products[p.ID] = p
	return nil
}

func (f *fakeStore) Get(_ context.Context, id string) (*catalog.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, catalog.ErrNotFound
	}
	return &p, nil
}

func (f *fakeStore) Search(_ context.Context, query string) ([]catalog.Product, error) {
	if f.searchFn != nil {
		return f.searchFn(query), nil
	}
	var results []catalog.Product
	for _, p := range f.products {
		results = append(results, p)
	}
	return results, nil
}

func TestCreateProduct(t *testing.T) {
	store := newFakeStore()
	h := New(store)

	body := `{"name":"Kulaklık","description":"Bluetooth","price":149.9,"stock":10}`
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.CreateProduct(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created catalog.Product
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated product ID, got empty string")
	}
	if created.Name != "Kulaklık" {
		t.Fatalf("expected name to round-trip, got %q", created.Name)
	}
	if _, ok := store.products[created.ID]; !ok {
		t.Fatal("expected product to be persisted in the store")
	}
}

func TestCreateProduct_MissingName(t *testing.T) {
	h := New(newFakeStore())

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"price":10}`))
	rec := httptest.NewRecorder()

	h.CreateProduct(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	h := New(newFakeStore())

	req := httptest.NewRequest(http.MethodGet, "/products/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	h.GetProduct(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetProduct_Found(t *testing.T) {
	store := newFakeStore()
	store.products["p1"] = catalog.Product{ID: "p1", Name: "Klavye", Price: 499, Stock: 3}
	h := New(store)

	req := httptest.NewRequest(http.MethodGet, "/products/p1", nil)
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.GetProduct(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got catalog.Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Name != "Klavye" {
		t.Fatalf("expected name Klavye, got %q", got.Name)
	}
}

func TestSearchProducts_RequiresQuery(t *testing.T) {
	h := New(newFakeStore())

	req := httptest.NewRequest(http.MethodGet, "/products/search", nil)
	rec := httptest.NewRecorder()

	h.SearchProducts(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestSearchProducts_ReturnsMatches(t *testing.T) {
	store := newFakeStore()
	store.searchFn = func(query string) []catalog.Product {
		return []catalog.Product{{ID: "p1", Name: "Kulaklık " + query}}
	}
	h := New(store)

	req := httptest.NewRequest(http.MethodGet, "/products/search?q=bluetooth", nil)
	rec := httptest.NewRecorder()

	h.SearchProducts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var results []catalog.Product
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Kulaklık bluetooth" {
		t.Fatalf("unexpected search results: %+v", results)
	}
}

func TestHealth(t *testing.T) {
	h := New(newFakeStore())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
