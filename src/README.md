# src/ — 业务服务目录

每个服务一个子目录，由对应成员以 PR 交付（本目录当前不含服务实现，仅契约与约定）。

## 服务矩阵

| 目录 | 服务 | 语言 | 契约（protos/） | 消费的客户端 | 默认端口 | 健康检查服务名 |
| --- | --- | --- | --- | --- | --- | --- |
| src/frontend | frontend | Go | 无（纯客户端） | 全部服务 | 8080 (HTTP) | — |
| src/checkoutservice | checkoutservice | Go | checkout.proto | cart/currency/product_catalog/shipping/email/payment | 5050 | hipstershop.CheckoutService |
| src/productcatalogservice | productcatalogservice | Go | product_catalog.proto | — | 3550 | hipstershop.ProductCatalogService |
| src/shippingservice | shippingservice | Go | shipping.proto | — | 50051 | hipstershop.ShippingService |
| src/emailservice | emailservice | Python | email.proto | — | 8080 | hipstershop.EmailService |
| src/recommendationservice | recommendationservice | Python | recommendation.proto | product_catalog | 8080 | hipstershop.RecommendationService |
| src/shoppingassistantservice | shoppingassistantservice | Python | shopping_assistant.proto | product_catalog | 8082 | hipstershop.ShoppingAssistantService |
| src/loadgenerator | loadgenerator | Python | 无（纯客户端） | 全部服务 | — | — |
| src/currencyservice | currencyservice | Node.js | currency.proto | — | 7000 | hipstershop.CurrencyService |
| src/paymentservice | paymentservice | Node.js | payment.proto | — | 50051 | hipstershop.PaymentService |
| src/adservice | adservice | Java (Spring Boot + Gradle) | ad.proto | — | 9555 | hipstershop.AdService |
| src/cartservice | cartservice | C# (.NET + Redis) | cart.proto | — | 7070 | hipstershop.CartService |

## 使用生成代码

- **Go 服务**：若为独立 go module，在服务 `go.mod` 中加入
  `replace github.com/Ayanamiz41/microservice-demo => ../..`，再
  `import "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"`。
- **Python 服务**：将 `genproto/python` 加入 `sys.path` 后 `import demo_pb2, cart_pb2_grpc, ...`。
- **Node.js 服务**：`require('../../genproto/nodejs/cart_pb.js')` 等；运行时依赖 `@grpc/grpc-js`、`google-protobuf`。
- **Java 服务**：把 `genproto/java` 加入源码目录或作为生成的 sources 使用；运行时依赖 grpc-java + protobuf-java。
- **C# 服务**：把 `genproto/csharp` 的 `.cs` 加入项目（或由项目内 Grpc.Tools 按需重新生成）。

## 交付验收（每个服务 PR）

1. 本地启动命令见各服务 README；
2. 单元测试全绿；
3. `grpcurl -plaintext -d '{"service": "<服务名>"}' localhost:<端口> grpc.health.v1.Health/Check` 返回 `SERVING`。
