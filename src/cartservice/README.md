# cartservice（C# / .NET 8 + Redis）

`hipstershop.CartService`（AddItem / GetCart / EmptyCart）的本地复刻，行为对齐上游
GoogleCloudPlatform/microservices-demo 的 cartservice。购物车存储在 Redis
（每用户一个 hash，key `cart:<userId>`，field=productId，value=quantity）；
Redis 不可用时自动回退线程安全的内存存储，可无依赖启动。

## 启动

```bash
dotnet run --project src/cartservice
```

- 默认端口 **7070**，可用环境变量覆盖：

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `PORT` | `7070` | 监听端口 |
| `CART_STORE` | `redis` | `redis` \| `memory`；`memory` 强制内存存储 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址（`CART_STORE=redis` 时使用；连不上则回退内存存储） |

## 测试

```bash
dotnet test src/cartservice/tests
```

冒烟测试通过 TestServer 宿主真实服务 + 生成客户端驱动（AddItem/GetCart/EmptyCart、
参数校验、gRPC health check），不依赖外部 Redis。

## gRPC health check

```bash
grpcurl -plaintext -d '{"service": "hipstershop.CartService"}' localhost:7070 grpc.health.v1.Health/Check
# 期望输出 {"status":"SERVING"}
```

## 代码结构

```
src/cartservice/
├── Program.cs            # 宿主、端口、存储选择（Redis / 内存回退）
├── CartServiceImpl.cs    # hipstershop.CartService 实现
├── HealthServiceImpl.cs  # grpc.health.v1.Health 实现
├── ICartStore.cs         # 存储抽象
├── RedisCartStore.cs     # Redis 存储（StackExchange.Redis）
├── LocalCartStore.cs     # 内存存储（线程安全，无依赖启动）
└── tests/                # xunit 冒烟测试（dotnet test）
```

生成的 C# 桩代码来自根目录 `genproto/csharp/`（已提交，勿手改），通过
`Compile Include="../../genproto/csharp/*.cs"` 链接进项目。
