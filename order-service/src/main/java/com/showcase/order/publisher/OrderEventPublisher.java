package com.showcase.order.publisher;

import com.showcase.order.config.RabbitMQConfig;
import com.showcase.order.model.Order;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

@Component
public class OrderEventPublisher {

    private final RabbitTemplate rabbitTemplate;

    public OrderEventPublisher(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void publishOrderCreated(Order order) {
        List<Map<String, Object>> items = order.getItems().stream()
                .map(item -> Map.<String, Object>of(
                        "productId", item.getProductId(),
                        "quantity", item.getQuantity()))
                .toList();

        Map<String, Object> event = Map.of(
                "orderId", order.getId(),
                "customerId", order.getCustomerId(),
                "items", items);

        rabbitTemplate.convertAndSend(RabbitMQConfig.EXCHANGE, RabbitMQConfig.ORDER_CREATED_ROUTING_KEY, event);
    }
}
