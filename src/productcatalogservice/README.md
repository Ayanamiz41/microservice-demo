# productcatalogservice

Go implementation of the Online Boutique **product catalog service** (KKS-28).

Serves the `hipstershop.ProductCatalogService` gRPC contract from
[`protos/product_catalog.proto`](../../protos/product_catalog.proto):

| RPC | Description |
| --- | --- |
| `ListProducts` | Returns the full product catalog. |
| `GetProduct(id)` | Returns a single product; `NotFound` if the id is unknown. |
| `SearchProducts(query)` | Case-insensitive substring match on name/description. |

The catalog data is built into the binary (`products.json`, embedded via
`go:embed`), mirroring the upstream Online Boutique product catalog.

## Prerequisites

- Go 1.25+ (`go.mod` at the repo root requires `go 1.25.0`)

## Local startup

```bash
# from the repo root
cd src/productcatalogservice
go run .
```

The server listens on `:3550` by default. Override with the `PORT`
environment variable:

```bash
PORT=3560 go run .
```

## Health check

The service registers the standard `grpc.health.v1.Health` service and reports
`SERVING` for the overall server and for
`hipstershop.ProductCatalogService`:

```bash
grpcurl -plaintext -d '{"service": "hipstershop.ProductCatalogService"}' \
  localhost:3550 grpc.health.v1.Health/Check
# → {"status": "SERVING"}
```

## Manual smoke test

```bash
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts
grpcurl -plaintext -d '{"id": "OLJCESPC7Z"}' localhost:3550 hipstershop.ProductCatalogService/GetProduct
grpcurl -plaintext -d '{"query": "kitchen"}' localhost:3550 hipstershop.ProductCatalogService/SearchProducts
```

## Unit tests

```bash
cd src/productcatalogservice
go test ./...
```
