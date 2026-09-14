package com.showcase.consumer.repository;

import com.showcase.consumer.model.Event;
import org.springframework.data.elasticsearch.repository.ElasticsearchRepository;

public interface EventSearchRepository extends ElasticsearchRepository<Event, String> {

    Iterable<Event> findByTypeContainingOrPayloadContaining(String type, String payload);
}
