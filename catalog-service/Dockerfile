FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/catalog-service .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/catalog-service /catalog-service
EXPOSE 8080
ENTRYPOINT ["/catalog-service"]
