package com.showcase.order.service;

import com.showcase.order.client.CatalogClient;
import com.showcase.order.client.ProductDto;
import com.showcase.order.dto.CreateOrderRequest;
import com.showcase.order.dto.OrderItemRequest;
import com.showcase.order.model.Order;
import com.showcase.order.model.OrderStatus;
import com.showcase.order.publisher.OrderEventPublisher;
import com.showcase.order.repository.OrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderServiceTest {

    @Mock
    private CatalogClient catalogClient;

    @Mock
    private OrderRepository orderRepository;

    @Mock
    private OrderEventPublisher eventPublisher;

    private OrderService orderService;

    @BeforeEach
    void setUp() {
        orderService = new OrderService(catalogClient, orderRepository, eventPublisher);
    }

    private ProductDto product(String id, String name, double price, int stock) {
        ProductDto dto = new ProductDto();
        dto.setId(id);
        dto.setName(name);
        dto.setPrice(price);
        dto.setStock(stock);
        return dto;
    }

    private CreateOrderRequest request(String customerId, String productId, int quantity) {
        OrderItemRequest item = new OrderItemRequest();
        item.setProductId(productId);
        item.setQuantity(quantity);

        CreateOrderRequest req = new CreateOrderRequest();
        req.setCustomerId(customerId);
        req.setItems(List.of(item));
        return req;
    }

    @Test
    void placeOrder_computesTotalAndPublishesEvent() {
        when(catalogClient.findProduct("p1")).thenReturn(Optional.of(product("p1", "Kulaklık", 100.0, 10)));
        when(orderRepository.save(any(Order.class))).thenAnswer(invocation -> invocation.getArgument(0));

        Order order = orderService.placeOrder(request("cust-1", "p1", 3));

        assertThat(order.getId()).isNotBlank();
        assertThat(order.getStatus()).isEqualTo(OrderStatus.CREATED);
        assertThat(order.getTotal()).isEqualTo(300.0);
        assertThat(order.getItems()).hasSize(1);
        assertThat(order.getItems().get(0).getProductName()).isEqualTo("Kulaklık");

        verify(orderRepository).save(order);
        verify(eventPublisher).publishOrderCreated(order);
    }

    @Test
    void placeOrder_rejectsUnknownProduct() {
        when(catalogClient.findProduct("missing")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> orderService.placeOrder(request("cust-1", "missing", 1)))
                .isInstanceOf(ProductNotFoundException.class);

        verifyNoInteractions(orderRepository, eventPublisher);
    }

    @Test
    void placeOrder_rejectsInsufficientStock() {
        when(catalogClient.findProduct("p1")).thenReturn(Optional.of(product("p1", "Kulaklık", 100.0, 1)));

        assertThatThrownBy(() -> orderService.placeOrder(request("cust-1", "p1", 5)))
                .isInstanceOf(InsufficientStockException.class);

        verifyNoInteractions(orderRepository, eventPublisher);
    }

    @Test
    void getOrder_returnsStoredOrder() {
        Order stored = new Order("o1", "cust-1", List.of(), 50.0, OrderStatus.CREATED, null);
        when(orderRepository.findById("o1")).thenReturn(Optional.of(stored));

        assertThat(orderService.getOrder("o1")).isSameAs(stored);
    }

    @Test
    void getOrder_throwsWhenMissing() {
        when(orderRepository.findById("missing")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> orderService.getOrder("missing"))
                .isInstanceOf(OrderNotFoundException.class);
    }
}
