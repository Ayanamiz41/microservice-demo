# microservice-demo（Online Boutique 复刻）

复刻 GoogleCloudPlatform/microservices-demo（Online Boutique）的 monorepo。所有服务**本地启动**，不包含任何 Docker / k8s / 云部署文件。

## 服务清单（12 个）

| 服务 | 语言/栈 | 契约文件 | 默认端口 |
| --- | --- | --- | --- |
| frontend | Go | 纯客户端（消费全部服务） | 8080 (HTTP) |
| checkoutservice | Go | `checkout.proto` | 5050 |
| productcatalogservice | Go | `product_catalog.proto` | 3550 |
| shippingservice | Go | `shipping.proto` | 50051 |
| emailservice | Python | `email.proto` | 8080 |
| recommendationservice | Python | `recommendation.proto` | 8080 |
| shoppingassistantservice | Python | `shopping_assistant.proto` | 8082 |
| loadgenerator | Python | 纯客户端 | — |
| currencyservice | Node.js | `currency.proto` | 7000 |
| paymentservice | Node.js | `payment.proto` | 50051 |
| adservice | Java (Spring Boot + Gradle) | `ad.proto` | 9555 |
| cartservice | C# (.NET + Redis) | `cart.proto` | 7070 |

> 端口沿用上游默认值；同一台机器本地调试时注意 8080 / 50051 端口占用（emailservice 与 recommendationservice 上游均为 8080，分属独立部署，本地逐个启动互不冲突）。

## 目录结构

```
├── protos/                 # gRPC/protobuf 契约（唯一事实来源）
│   ├── demo.proto          # 共享领域消息（Money/Address/CartItem/Product/OrderResult…）
│   ├── <service>.proto     # 每个 gRPC 服务一份契约
│   └── grpc/health/v1/health.proto   # 标准健康检查契约
├── genproto/               # 已提交的生成代码（勿手改）
│   ├── go/                 # Go（genproto/go/hipstershop）
│   ├── python/             # Python（*_pb2.py / *_pb2_grpc.py）
│   ├── nodejs/             # Node.js（*_pb.js / *_grpc_pb.js，@grpc/grpc-js 风格）
│   ├── java/               # Java（hipstershop/*.java）
│   └── csharp/             # C#（*.cs）
├── genproto.sh             # 重新生成脚本（bash / Git Bash / Linux / macOS）
├── genproto.ps1            # 重新生成脚本（Windows PowerShell）
├── go.mod                  # 根 Go module（使 genproto/go 可直接 go build ./...）
└── src/                    # 各业务服务代码（README.md 见服务矩阵）
```

## 契约（protos/）

- 与上游 **wire-compatible**：同一 `package hipstershop`、同一服务/方法/消息名与字段号；`demo.proto` 仅保留共享领域消息，服务定义按服务拆分到各自的 `.proto`。
- `frontend`、`loadgenerator` 为纯客户端服务，无服务端契约，直接复用其余服务的生成客户端。
- `shoppingassistantservice` 上游为纯云 Flask 应用（AlloyDB/Gemini），无 gRPC 契约；本复刻按本地范围简化为 `ShoppingAssistantService.GetCompletion`。
- 健康检查使用标准 `grpc.health.v1.Health`（见 `protos/grpc/health/v1/health.proto`）。

## 重新生成代码（genproto）

修改任何 `.proto` 后必须重新生成并提交 `genproto/` 产物：

```bash
# Linux / macOS / Git Bash
./genproto.sh

# Windows PowerShell
.\genproto.ps1
```

工具要求（全部在 PATH 上，或通过环境变量 `PROTOC`/`PYTHON` 指定）：

| 工具 | 用途 | 安装 |
| --- | --- | --- |
| protoc (≥3.19) | 驱动 | `choco install protobuf` 等 |
| protoc-gen-go | Go messages | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| protoc-gen-go-grpc | Go services | `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest` |
| python + grpcio-tools | Python | `pip install grpcio-tools` |
| grpc_tools_node_protoc | Node.js | `npm install -g grpc-tools` |
| protoc-gen-grpc-java | Java | Maven Central `io.grpc:protoc-gen-grpc-java` |
| protoc-gen-grpc_csharp | C# | NuGet `Grpc.Tools`（tools/windows_x64/） |

有意排除（见脚本内注释）：Go 与 Python 不生成 `grpc/health/v1`——Go 使用标准包 `google.golang.org/grpc/health/grpc_health_v1`；Python 使用 `grpcio-health-checking`（顶层 `grpc_health.v1`），避免本地 `grpc/` 目录遮蔽 grpc 运行时。

## 本地启动与验证（各服务）

各业务服务在 `src/<service>/` 下实现，由对应成员以 PR 交付；每个服务 PR 需附：启动命令、单测、gRPC health check 通过。

通用验证方法（服务启动后）：

```bash
# 安装 grpcurl 后，对任意服务做健康检查
grpcurl -plaintext -d '{"service": "hipstershop.CartService"}' localhost:7070 grpc.health.v1.Health/Check
```

## 协作规则

- 所有变更走 feature 分支 + PR；分支名与标题含 issue 编号（如 `KKS-18`），PR 描述含 `Closes KKS-18`。
- PR 由 QA & Integration 复核并合并；不直接推 main。
- 任何服务新增/修改接口：先更新 `protos/` 并重新生成 `genproto/`，随 PR 一起合入。
