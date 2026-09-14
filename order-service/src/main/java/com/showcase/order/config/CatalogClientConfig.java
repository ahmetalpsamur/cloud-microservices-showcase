package com.showcase.order.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

import java.time.Duration;

@Configuration
public class CatalogClientConfig {

    @Bean
    public RestTemplate catalogRestTemplate(RestTemplateBuilder builder,
                                             @Value("${catalog.service.url}") String catalogServiceUrl) {
        return builder
                .rootUri(catalogServiceUrl)
                .connectTimeout(Duration.ofSeconds(3))
                .readTimeout(Duration.ofSeconds(3))
                .build();
    }
}
