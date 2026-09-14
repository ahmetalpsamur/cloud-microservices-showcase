package com.showcase.order.client;

import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;

import java.util.Optional;

@Component
public class CatalogClient {

    private final RestTemplate restTemplate;

    public CatalogClient(RestTemplate catalogRestTemplate) {
        this.restTemplate = catalogRestTemplate;
    }

    public Optional<ProductDto> findProduct(String productId) {
        try {
            return Optional.ofNullable(restTemplate.getForObject("/products/{id}", ProductDto.class, productId));
        } catch (RestClientException e) {
            return Optional.empty();
        }
    }
}
