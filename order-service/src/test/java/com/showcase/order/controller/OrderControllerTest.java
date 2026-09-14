package com.showcase.order.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.showcase.order.model.Order;
import com.showcase.order.model.OrderItem;
import com.showcase.order.model.OrderStatus;
import com.showcase.order.service.InsufficientStockException;
import com.showcase.order.service.OrderNotFoundException;
import com.showcase.order.service.OrderService;
import com.showcase.order.service.ProductNotFoundException;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.time.Instant;
import java.util.List;
import java.util.Map;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(OrderController.class)
class OrderControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private OrderService orderService;

    @Test
    void createOrder_returns201WithBody() throws Exception {
        Order order = new Order("o1", "cust-1",
                List.of(new OrderItem("p1", "Kulaklık", 2, 100.0)),
                200.0, OrderStatus.CREATED, Instant.parse("2026-01-01T00:00:00Z"));
        when(orderService.placeOrder(any())).thenReturn(order);

        String body = objectMapper.writeValueAsString(Map.of(
                "customerId", "cust-1",
                "items", List.of(Map.of("productId", "p1", "quantity", 2))));

        mockMvc.perform(post("/orders").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.id").value("o1"))
                .andExpect(jsonPath("$.total").value(200.0));
    }

    @Test
    void createOrder_missingCustomerId_returns400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of(
                "items", List.of(Map.of("productId", "p1", "quantity", 2))));

        mockMvc.perform(post("/orders").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isBadRequest());
    }

    @Test
    void createOrder_productNotFound_returns422() throws Exception {
        when(orderService.placeOrder(any())).thenThrow(new ProductNotFoundException("p1"));

        String body = objectMapper.writeValueAsString(Map.of(
                "customerId", "cust-1",
                "items", List.of(Map.of("productId", "p1", "quantity", 2))));

        mockMvc.perform(post("/orders").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isUnprocessableEntity());
    }

    @Test
    void createOrder_insufficientStock_returns409() throws Exception {
        when(orderService.placeOrder(any())).thenThrow(new InsufficientStockException("not enough stock"));

        String body = objectMapper.writeValueAsString(Map.of(
                "customerId", "cust-1",
                "items", List.of(Map.of("productId", "p1", "quantity", 99))));

        mockMvc.perform(post("/orders").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isConflict());
    }

    @Test
    void getOrder_notFound_returns404() throws Exception {
        when(orderService.getOrder(eq("missing"))).thenThrow(new OrderNotFoundException("missing"));

        mockMvc.perform(get("/orders/missing"))
                .andExpect(status().isNotFound());
    }
}
