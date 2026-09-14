# Cloud Microservices Showcase

Küçük ama uçtan uca çalışan bir mikroservis mimarisi: **Go** tabanlı bir ingestion servisi, olayları **RabbitMQ** üzerinden **Spring Boot** tabanlı bir consumer servisine iletir, bu servis olayları **Elasticsearch**'e indeksler ve arama uç noktası sunar. Tamamı **Kubernetes** üzerinde çalışacak şekilde paketlenmiş, **AWS (EKS)** üzerinde Terraform ile provision edilebilir.

## Mimari

```
        POST /events                 events.exchange           events.queue
Client ───────────────▶ [go-service] ───────────────▶ RabbitMQ ───────────────▶ [spring-boot-service]
                          (Go, :8080)                                              (Java, :8081)
                                                                                        │
                                                                                        ▼
                                                                                  Elasticsearch
                                                                                        ▲
                                                                                        │
Client ◀────────────────────────────── GET /search?q=... ─────────────────────────────┘
```

## Kullanılan teknolojiler

| Alan | Teknoloji |
|---|---|
| Bulut | AWS (EKS) — `infra/terraform/aws` |
| Altyapı & orkestrasyon | Kubernetes — `k8s/` |
| Mesajlaşma | RabbitMQ — `go-service` publisher, `spring-boot-service` consumer |
| Arama/veri | Elasticsearch — `spring-boot-service` |
| Framework | Spring Boot — `spring-boot-service` |
| Dil | Go — `go-service` |

## Servisler

### `go-service`
- `POST /events` — gelen JSON olayını doğrular ve RabbitMQ'daki `events.exchange` exchange'ine publish eder.
- `GET /healthz` — health check.

### `spring-boot-service`
- RabbitMQ `events.queue` kuyruğunu dinler, gelen olayları Elasticsearch `events` index'ine yazar.
- `GET /search?q=...` — Elasticsearch üzerinde arama yapar.
- `GET /actuator/health` — health check.

## Lokal geliştirme

```bash
docker compose up --build
```

Bu komut RabbitMQ, Elasticsearch, `go-service` ve `spring-boot-service`'i ayağa kaldırır.

```bash
# Olay gönder
curl -X POST localhost:8080/events -d '{"type":"order.created","payload":{"orderId":"123"}}' -H 'Content-Type: application/json'

# Ara
curl 'localhost:8081/search?q=order'
```

## Kubernetes'e deploy

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/
```

## AWS altyapısı (Terraform)

`infra/terraform/aws` altında bir VPC ve EKS cluster tanımı bulunur:

```bash
cd infra/terraform/aws
terraform init
terraform plan
```

## CI

`.github/workflows/ci.yml` her push'ta Go servisini derler/test eder ve Spring Boot servisini Maven ile build eder.
