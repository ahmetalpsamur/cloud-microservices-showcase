package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrdersExchange = "orders.exchange"
	StockQueue     = "catalog.stock-sync"
	RoutingKey     = "order.created"
)

type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type OrderCreated struct {
	OrderID    string      `json:"orderId"`
	CustomerID string      `json:"customerId"`
	Items      []OrderItem `json:"items"`
}

// StockSyncHandler decrements catalog stock for every item in an order.
type StockSyncHandler func(ctx context.Context, productID string, quantity int) error

// Consumer listens for order.created events and applies stock updates.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(OrdersExchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(StockQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err := ch.QueueBind(StockQueue, RoutingKey, OrdersExchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch}, nil
}

// Run consumes messages until the context is cancelled.
func (c *Consumer) Run(ctx context.Context, handle StockSyncHandler) error {
	msgs, err := c.channel.Consume(StockQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var evt OrderCreated
			if err := json.Unmarshal(msg.Body, &evt); err != nil {
				log.Printf("discarding malformed order.created message: %v", err)
				msg.Nack(false, false)
				continue
			}

			failed := false
			for _, item := range evt.Items {
				if err := handle(ctx, item.ProductID, item.Quantity); err != nil {
					log.Printf("failed to sync stock for order=%s product=%s: %v", evt.OrderID, item.ProductID, err)
					failed = true
				}
			}

			if failed {
				msg.Nack(false, true)
				continue
			}
			log.Printf("synced stock for order=%s (%d items)", evt.OrderID, len(evt.Items))
			msg.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
