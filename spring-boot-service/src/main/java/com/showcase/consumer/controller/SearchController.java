package com.showcase.consumer.controller;

import com.showcase.consumer.model.Event;
import com.showcase.consumer.repository.EventSearchRepository;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SearchController {

    private final EventSearchRepository repository;

    public SearchController(EventSearchRepository repository) {
        this.repository = repository;
    }

    @GetMapping("/search")
    public Iterable<Event> search(@RequestParam("q") String query) {
        return repository.findByTypeContainingOrPayloadContaining(query, query);
    }
}
