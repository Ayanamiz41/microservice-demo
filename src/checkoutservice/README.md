# checkoutservice (Go)

Online Boutique 复刻项目的 CheckoutService，对齐上游
[GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)
`src/checkoutservice` 实现：`PlaceOrder` → `OrderResult` 全链路
（cart → product catalog + currency → shipping quote → payment → shipping
→ empty cart → email confirmation）。

- 契约：`protos/checkout.proto`（wire-compatible with upstream）
- 生成桩：`genproto/go/hipstershop`（官方 protoc-gen-go，不自写 gRPC 桩）
- 金额计算：`money` 包，纯整数（units + nanos），禁止浮点

## 本地启动

独立 Go module，在服务目录下运行：

```bash
cd src/checkoutservice
go run ./...
```

默认监听 `:5050`（`PORT` 可覆盖）。下游服务地址默认取
`src/README.md` 端口矩阵的本地值，无需配置即可启动；部署时用环境变量覆盖：

| 环境变量 | 默认值 | 下游服务 |
| --- | --- | --- |
| `PRODUCT_CATALOG_SERVICE_ADDR` | `localhost:3550` | productcatalogservice |
| `CART_SERVICE_ADDR` | `localhost:7070` | cartservice |
| `CURRENCY_SERVICE_ADDR` | `localhost:7000` | currencyservice |
| `SHIPPING_SERVICE_ADDR` | `localhost:50051` | shippingservice |
| `EMAIL_SERVICE_ADDR` | `localhost:8080` | emailservice |
| `PAYMENT_SERVICE_ADDR` | `localhost:50051` | paymentservice |

连接是惰性的（`grpc.NewClient`），启动不依赖下游在线。

## 测试

```bash
cd src/checkoutservice
go test ./...
```

- `money`：上游 money 包全量单测（整数金额运算）
- `server`：内存假下游（cart/product catalog/currency/shipping/payment/email）
  端到端跑通 `PlaceOrder`，校验 OrderResult、实扣金额（$15.00 = 2×$1.50 +
  $2.00 + $10.00 运费）、空购物车与邮件确认；另含 gRPC health check 用例

## 健康检查

```bash
grpcurl -plaintext -d '{"service":"hipstershop.CheckoutService"}' \
  localhost:5050 grpc.health.v1.Health/Check
# → {"status":"SERVING"}
```
