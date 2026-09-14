package com.showcase.order.service;

public class OrderNotFoundException extends RuntimeException {

    public OrderNotFoundException(String orderId) {
        super("order not found: " + orderId);
    }
}
