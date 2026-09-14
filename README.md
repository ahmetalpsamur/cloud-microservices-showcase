# Cloud Microservices Showcase — E-Ticaret Sipariş Akışı

Tek bir iş akışına odaklanan, uçtan uca çalışan bir e-ticaret backend'i: ürün kataloğu ve arama **Go + Elasticsearch** ile, sipariş verme **Spring Boot** ile yapılır; iki servis **RabbitMQ** üzerinden haberleşir. Tamamı **Kubernetes**'te çalışacak şekilde paketlenmiş, **AWS (EKS)** üzerinde Terraform ile provision edilebilir.

## İş akışı

1. Ürünler `catalog-service`'e eklenir ve Elasticsearch'e indekslenir.
2. Müşteri `catalog-service` üzerinden ürün arar (`GET /products/search?q=`).
3. Müşteri `order-service`'e sipariş verir (`POST /orders`). Sipariş kalemleri `catalog-service`'ten fiyat/stok bilgisiyle doğrulanır; stok yetersizse sipariş reddedilir.
4. `order-service` siparişi kaydeder ve `order.created` olayını RabbitMQ'nun `orders.exchange`'ine yayınlar.
5. `catalog-service`, `order.created` olayını dinler ve ilgili ürünlerin stoğunu Elasticsearch'te günceller — böylece bir sonraki arama güncel stoğu yansıtır.

```
                 POST /products                    GET /products/search
Admin ─────────────────────────▶ [catalog-service] ◀───────────────────── Müşteri
                                    (Go, :8080)
                                    Elasticsearch
                                        ▲   │
                          order.created │   │ GET /products/{id}
                       (stok senkronu)  │   │ (fiyat/stok doğrulama)
                                        │   ▼
                                    RabbitMQ ◀── order.created yayınla ── [order-service]
                                                                            (Java, :8081)
                                                                                 ▲
                                                                          POST /orders
                                                                                 │
                                                                             Müşteri
```

## Kullanılan teknolojiler

| Alan | Teknoloji |
|---|---|
| Bulut | AWS (EKS) — `infra/terraform/aws` |
| Altyapı & orkestrasyon | Kubernetes — `k8s/` |
| Mesajlaşma | RabbitMQ — sipariş → stok senkronu |
| Arama/veri | Elasticsearch — ürün kataloğu ve arama |
| Framework | Spring Boot — `order-service` |
| Dil | Go — `catalog-service` |

## Servisler

### `catalog-service` (Go, :8080)
- `POST /products` — ürün oluşturur, Elasticsearch `products` index'ine yazar.
- `GET /products/{id}` — tek ürünü getirir.
- `GET /products/search?q=` — ürün adı/açıklamasında tam metin arama yapar.
- `GET /healthz` — health check.
- Arka planda: `order.created` olaylarını RabbitMQ'dan dinleyip ilgili ürünlerin stoğunu düşürür.

### `order-service` (Spring Boot, :8081)
- `POST /orders` — sipariş kalemlerini `catalog-service`'ten doğrular (fiyat + stok), toplamı hesaplar, siparişi kaydeder ve `order.created` olayını yayınlar.
- `GET /orders/{id}` — sipariş durumunu döner.
- `GET /actuator/health` — health check.

> Not: Siparişler bu demo'da bellek içi (in-memory) tutulur; gerçek bir kurulumda bir veritabanının (ör. PostgreSQL) arkasına alınması beklenir.

## Lokal geliştirme

```bash
docker compose up --build
```

```bash
# Ürün ekle
curl -X POST localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Kablosuz Kulaklık","description":"Gürültü engellemeli bluetooth kulaklık","price":149.90,"stock":25}'

# Ara
curl 'localhost:8080/products/search?q=kulaklık'

# Sipariş ver (dönen ürün id'sini kullan)
curl -X POST localhost:8081/orders \
  -H 'Content-Type: application/json' \
  -d '{"customerId":"cust-1","items":[{"productId":"<product-id>","quantity":2}]}'

# Sipariş sonrası stoğun düştüğünü doğrula
curl 'localhost:8080/products/search?q=kulaklık'
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

`.github/workflows/ci.yml` her push'ta `catalog-service`'i derler/vet eder, `order-service`'i Maven ile build eder ve Kubernetes manifestlerini `kubeconform` ile doğrular.
