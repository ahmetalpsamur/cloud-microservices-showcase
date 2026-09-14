# Cloud Microservices Showcase — E-Ticaret Sipariş Akışı

[![CI](https://github.com/ahmetalpsamur/cloud-microservices-showcase/actions/workflows/ci.yml/badge.svg)](https://github.com/ahmetalpsamur/cloud-microservices-showcase/actions/workflows/ci.yml)

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

Siparişler **PostgreSQL**'de tutulur (Spring Data JPA + Flyway ile şema yönetimi, `order-service/src/main/resources/db/migration`).

## Test

```bash
cd catalog-service && go test ./...
cd order-service && mvn test
```

CI her push'ta bu testleri, Kubernetes manifest lint'ini ve (main'e push'ta) Docker image build+publish adımını çalıştırır.

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

Servis imajları CI tarafından her main push'unda `ghcr.io/ahmetalpsamur/{catalog-service,order-service}:latest` olarak build edilip yayınlanır (bkz. `.github/workflows/ci.yml`).

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/
```

> Not: `k8s/postgres.yaml` demo amaçlı `emptyDir` volume kullanır (pod yeniden başlarsa veri kaybolur). Gerçek bir kurulumda bunun yerine bir `PersistentVolumeClaim` veya yönetilen bir veritabanı (RDS vb.) kullanılmalı.

## AWS altyapısı (Terraform)

`infra/terraform/aws` altında bir VPC ve EKS cluster tanımı bulunur. State S3'te, kilitleme DynamoDB'de tutulur; bucket/tablo `infra/terraform/bootstrap` ile bir kere oluşturulur:

```bash
# 1) State backend'i bir kere provision et (kendi local state'iyle çalışır)
cd infra/terraform/bootstrap
terraform init
terraform apply -var="state_bucket_name=<globally-unique-bucket-name>"

# 2) Asıl altyapıyı, oluşan bucket/tabloyu backend olarak kullanarak init et
cd ../aws
terraform init \
  -backend-config="bucket=<globally-unique-bucket-name>" \
  -backend-config="key=cloud-microservices-showcase/terraform.tfstate" \
  -backend-config="region=eu-central-1" \
  -backend-config="dynamodb_table=cloud-microservices-showcase-tflock"
terraform plan
```

## CI

`.github/workflows/ci.yml`:
- `catalog-service`'i derler, vet eder ve test eder (Go)
- `order-service`'i Maven ile build edip test eder (Spring Boot)
- Kubernetes manifestlerini `kubeconform` ile doğrular
- main'e her push'ta her iki servisin Docker imajını build edip GHCR'a yayınlar (`:latest` ve commit SHA'sı ile)
