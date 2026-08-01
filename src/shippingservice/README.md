# shippingservice (Go)

`hipstershop.ShippingService` 的 Go 实现，复刻 GoogleCloudPlatform/microservices-demo
（Online Boutique）的对应服务：

- `GetQuote`：按购物车商品总件数报价 —— 空购物车免费，否则固定 USD 8.99
  （金额计算全部使用整数最小单位，不涉及浮点）；
- `ShipOrder`：按收货地址生成模拟 tracking id。

- 契约：`protos/shipping.proto`（wire-compatible with upstream）
- 生成代码：`genproto/go/hipstershop`
- 端口：默认 `50051`（环境变量 `PORT` 可覆盖）
- 健康检查服务名：`hipstershop.ShippingService`

## 本地启动

```bash
cd src/shippingservice
go run .
```

## 单元测试

```bash
cd src/shippingservice
go test ./...
```

## gRPC health check

```bash
grpcurl -plaintext -d '{"service": "hipstershop.ShippingService"}' \
  localhost:50051 grpc.health.v1.Health/Check
# 期望返回 {"status":"SERVING"}
```

## 冒烟测试

```bash
# GetQuote（4 件商品 → 8.99 USD）
grpcurl -plaintext -d '{
  "address": {"streetAddress": "Muffin Man", "city": "London", "country": "England"},
  "items": [{"productId": "23", "quantity": 1}, {"productId": "46", "quantity": 3}]
}' localhost:50051 hipstershop.ShippingService/GetQuote

# ShipOrder（返回 18 位 tracking id）
grpcurl -plaintext -d '{
  "address": {"streetAddress": "Muffin Man", "city": "London", "country": "England"},
  "items": [{"productId": "23", "quantity": 1}]
}' localhost:50051 hipstershop.ShippingService/ShipOrder
```
