package com.showcase.consumer.listener;

import com.showcase.consumer.config.RabbitMQConfig;
import com.showcase.consumer.model.Event;
import com.showcase.consumer.repository.EventSearchRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

@Component
public class EventListener {

    private static final Logger log = LoggerFactory.getLogger(EventListener.class);

    private final EventSearchRepository repository;

    public EventListener(EventSearchRepository repository) {
        this.repository = repository;
    }

    @RabbitListener(queues = RabbitMQConfig.QUEUE)
    public void onEvent(Map<String, Object> message) {
        String type = String.valueOf(message.getOrDefault("type", "unknown"));
        Object payload = message.getOrDefault("payload", "{}");

        Event event = new Event(UUID.randomUUID().toString(), type, String.valueOf(payload), Instant.now());
        repository.save(event);

        log.info("Indexed event type={} id={}", type, event.getId());
    }
}
