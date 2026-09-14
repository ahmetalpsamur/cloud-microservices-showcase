package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ErrNotFound is returned when a product does not exist in the index.
var ErrNotFound = fmt.Errorf("product not found")

// Store is a thin client over the Elasticsearch REST API that keeps the
// product catalog as documents in a single index. It intentionally avoids
// a full Elasticsearch SDK to keep the wire format explicit.
type Store struct {
	baseURL string
	index   string
	http    *http.Client
}

func NewStore(baseURL, index string) *Store {
	return &Store{baseURL: baseURL, index: index, http: &http.Client{}}
}

// EnsureIndex creates the index with explicit mappings if it doesn't exist yet.
func (s *Store) EnsureIndex(ctx context.Context) error {
	if exists, err := s.do(ctx, http.MethodHead, "/"+s.index, nil); err == nil {
		exists.Body.Close()
		if exists.StatusCode == http.StatusOK {
			return nil
		}
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":          map[string]any{"type": "keyword"},
				"name":        map[string]any{"type": "text"},
				"description": map[string]any{"type": "text"},
				"price":       map[string]any{"type": "double"},
				"stock":       map[string]any{"type": "integer"},
			},
		},
	}
	resp, err := s.do(ctx, http.MethodPut, "/"+s.index, mapping)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("create index: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, p Product) error {
	resp, err := s.do(ctx, http.MethodPut, fmt.Sprintf("/%s/_doc/%s?refresh=true", s.index, p.ID), p)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("index product: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (*Product, error) {
	resp, err := s.do(ctx, http.MethodGet, fmt.Sprintf("/%s/_doc/%s", s.index, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get product: unexpected status %d", resp.StatusCode)
	}

	var doc struct {
		Source Product `json:"_source"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc.Source, nil
}

func (s *Store) Search(ctx context.Context, query string) ([]Product, error) {
	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"name^2", "description"},
			},
		},
	}

	resp, err := s.do(ctx, http.MethodPost, fmt.Sprintf("/%s/_search", s.index), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search products: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		products = append(products, hit.Source)
	}
	return products, nil
}

// DecrementStock atomically reduces the stock of a product by qty, clamped at zero.
func (s *Store) DecrementStock(ctx context.Context, id string, qty int) error {
	body := map[string]any{
		"script": map[string]any{
			"source": "ctx._source.stock = Math.max(0, ctx._source.stock - params.qty)",
			"lang":   "painless",
			"params": map[string]any{"qty": qty},
		},
	}
	resp, err := s.do(ctx, http.MethodPost, fmt.Sprintf("/%s/_update/%s?refresh=true", s.index, id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("decrement stock: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (s *Store) do(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.http.Do(req)
}
