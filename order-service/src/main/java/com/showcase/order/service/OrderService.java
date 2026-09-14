package com.showcase.order.service;

import com.showcase.order.client.CatalogClient;
import com.showcase.order.client.ProductDto;
import com.showcase.order.dto.CreateOrderRequest;
import com.showcase.order.dto.OrderItemRequest;
import com.showcase.order.model.Order;
import com.showcase.order.model.OrderItem;
import com.showcase.order.model.OrderStatus;
import com.showcase.order.publisher.OrderEventPublisher;
import com.showcase.order.repository.OrderRepository;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

@Service
public class OrderService {

    private final CatalogClient catalogClient;
    private final OrderRepository orderRepository;
    private final OrderEventPublisher eventPublisher;

    public OrderService(CatalogClient catalogClient, OrderRepository orderRepository,
                         OrderEventPublisher eventPublisher) {
        this.catalogClient = catalogClient;
        this.orderRepository = orderRepository;
        this.eventPublisher = eventPublisher;
    }

    public Order placeOrder(CreateOrderRequest request) {
        List<OrderItem> items = request.getItems().stream()
                .map(this::resolveItem)
                .toList();

        double total = items.stream().mapToDouble(OrderItem::lineTotal).sum();

        Order order = new Order(
                UUID.randomUUID().toString(),
                request.getCustomerId(),
                items,
                total,
                OrderStatus.CREATED,
                Instant.now());

        orderRepository.save(order);
        eventPublisher.publishOrderCreated(order);

        return order;
    }

    public Order getOrder(String id) {
        return orderRepository.findById(id)
                .orElseThrow(() -> new OrderNotFoundException(id));
    }

    private OrderItem resolveItem(OrderItemRequest itemRequest) {
        ProductDto product = catalogClient.findProduct(itemRequest.getProductId())
                .orElseThrow(() -> new ProductNotFoundException(itemRequest.getProductId()));

        if (product.getStock() < itemRequest.getQuantity()) {
            throw new InsufficientStockException(
                    "insufficient stock for product " + product.getId()
                            + ": requested " + itemRequest.getQuantity() + ", available " + product.getStock());
        }

        return new OrderItem(product.getId(), product.getName(), itemRequest.getQuantity(), product.getPrice());
    }
}
