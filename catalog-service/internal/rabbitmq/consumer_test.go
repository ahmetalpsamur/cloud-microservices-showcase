package rabbitmq

import (
	"encoding/json"
	"testing"
)

func TestOrderCreated_Unmarshal(t *testing.T) {
	payload := `{"orderId":"o1","customerId":"c1","items":[{"productId":"p1","quantity":2},{"productId":"p2","quantity":1}]}`

	var evt OrderCreated
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		t.Fatalf("failed to unmarshal order.created payload: %v", err)
	}

	if evt.OrderID != "o1" || evt.CustomerID != "c1" {
		t.Fatalf("unexpected order metadata: %+v", evt)
	}
	if len(evt.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(evt.Items))
	}
	if evt.Items[0].ProductID != "p1" || evt.Items[0].Quantity != 2 {
		t.Fatalf("unexpected first item: %+v", evt.Items[0])
	}
}
