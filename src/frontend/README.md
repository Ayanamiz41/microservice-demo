# frontend (Go)

Online Boutique 复刻项目的 **frontend** 服务，对齐上游
[GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)
`src/frontend` 实现：用 `templates/`（Go `html/template`）+ `static/` 渲染店铺
Web UI，并通过 gRPC 聚合全部后端服务（product catalog / cart / currency /
shipping / checkout / recommendations / ads / shopping assistant），覆盖
**浏览目录 → 购物车 → 结账 → 订单确认** 完整用户链路。

- **契约**：纯客户端，无服务端契约（`src/README.md` 矩阵中健康检查服务名为 "—"）
- **生成桩**：`genproto/go/hipstershop`（官方 protoc-gen-go，不自写 gRPC 桩）
- **金额计算**：`money` 包，纯整数（units + nanos），禁止浮点
- **路由**：与上游一致（`/`、`/product/{id}`、`/cart`、`/cart/checkout`、
  `/setCurrency`、`/logout`、`/assistant`、`/bot`、`/product-meta/{ids}`、
  `/static/`、`/_healthz`、`/robots.txt`），使用 Go 标准库 `net/http` mux

## 本地启动

独立 Go module，在服务目录下运行：

```bash
cd src/frontend
go run ./...
```

默认监听 `:8080`（`PORT` 覆盖端口，`LISTEN_ADDR` 覆盖监听地址）。下游服务
地址默认取 `src/README.md` 端口矩阵的本地值，无需配置即可启动；部署时用环境
变量覆盖：

| 环境变量 | 默认值 | 下游服务 |
| --- | --- | --- |
| `PRODUCT_CATALOG_SERVICE_ADDR` | `localhost:3550` | productcatalogservice |
| `CART_SERVICE_ADDR` | `localhost:7070` | cartservice |
| `CURRENCY_SERVICE_ADDR` | `localhost:7000` | currencyservice |
| `RECOMMENDATION_SERVICE_ADDR` | `localhost:8080` | recommendationservice |
| `CHECKOUT_SERVICE_ADDR` | `localhost:5050` | checkoutservice |
| `SHIPPING_SERVICE_ADDR` | `localhost:50051` | shippingservice |
| `AD_SERVICE_ADDR` | `localhost:9555` | adservice |
| `SHOPPING_ASSISTANT_SERVICE_ADDR` | `localhost:8082` | shoppingassistantservice |

连接是惰性的（`grpc.NewClient`），启动不依赖下游在线。对**非关键**服务做了与
上游一致的优雅降级：

- 广告 / 推荐不可达：仅记日志，页面照常渲染（上游行为）
- currencyservice 不可达：货币下拉回退到白名单（USD/EUR/CAD/JPY/GBP/TRY），
  默认 USD 流程因 `avoidNoopCurrencyConversionRPC` 短路不依赖汇率服务；
  切换非 USD 货币时才需要 currencyservice 在线

其余（目录 / 购物车 / 运费 / 结账）与上游一致，下游不可达时渲染错误页。

## 测试

```bash
cd src/frontend
go test ./...
```

- `money`：上游 money 包全量单测（整数金额运算，与 checkoutservice 同源）
- `validator`：表单校验（数量 1..10、必填字段、email、信用卡 Luhn 校验、
  ISO 4217 货币码）
- `server`：内存假下游（product catalog/cart/currency/recommendation/
  shipping/checkout/ad/assistant）端到端跑通全部路由：首页/商品页渲染、
  加购/清空购物车、购物车合计（整数金额：2×$19.99 + $10.00 运费 = $49.98）、
  货币切换、下单→订单确认、错误/校验错误页、`/product-meta` JSON、
  `/bot` 助手对话、会话 cookie、静态资源，以及 gRPC health check

## 健康检查

frontend 是 HTTP 服务，健康检查有两条等价途径，都在 `:8080` 单端口上：

```bash
# HTTP
curl http://localhost:8080/_healthz            # → ok

# gRPC health（与 grpcurl 同款 h2c 明文 HTTP/2 传输，同端口提供）
grpcurl -plaintext -d '{"service":"frontend"}' \
  localhost:8080 grpc.health.v1.Health/Check   # → {"status":"SERVING"}
```

## 手动冒烟（跨服务联通，可选）

```bash
cd src/productcatalogservice && go run ./... &   # :3550
cd src/checkoutservice && go run ./... &         # :5050
cd src/frontend && go run ./... &                # :8080
```

```bash
curl http://localhost:8080/_healthz
curl http://localhost:8080/product-meta/OLJCESPC7Z   # 经 productcatalog 返回 JSON
# 浏览器打开 http://localhost:8080/ 即可浏览（购物车/货币等页面需对应服务在线）
```
