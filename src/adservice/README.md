# AdService（广告推荐服务）

复刻 GoogleCloudPlatform/microservices-demo（Online Boutique）的 adservice：基于请求中的
上下文关键词（商品分类）返回广告；无匹配或空请求时返回随机广告。

- 语言/栈：Java 17 + Spring Boot 3.2 + Gradle
- 契约：`protos/ad.proto`（`hipstershop.AdService.GetAds`），直接消费已提交的
  `genproto/java` 生成代码（构建期无需 protoc）
- 默认端口：`9555`（可用环境变量 `PORT` 覆盖）
- 健康检查：`grpc.health.v1.Health`，服务名 `hipstershop.AdService`

## 环境要求

- JDK 17（`JAVA_HOME` 需指向 JDK 17，例如本机
  `C:\Program Files\Microsoft\jdk-17.0.11.9-hotspot`）
- 首次构建 Gradle wrapper 会自动下载 Gradle 8.5 发行版

## 本地启动

```bash
cd src/adservice
./gradlew bootRun          # Windows: .\gradlew.bat bootRun
```

启动后应看到 `Ad Service started, listening on 9555`。

## 单元测试

```bash
cd src/adservice
./gradlew test
```

## 健康检查

```bash
grpcurl -plaintext -d '{"service": "hipstershop.AdService"}' localhost:9555 grpc.health.v1.Health/Check
# => {"status": "SERVING"}
```

也可用任意 gRPC 客户端调用 `GetAds`：

```json
{"context_keys": ["clothing"]}
```

## 设计说明

- `AdServiceApplication`：Spring Boot 入口；
- `AdServiceImpl`：`GetAds` RPC 实现（对齐上游行为：按 context_keys 聚合、空/无匹配回退随机广告）；
- `AdCatalog`：内置广告目录（7 条广告、6 个分类），随机广告逻辑可注入种子便于测试；
- `HealthServiceImpl`：基于提交的 `genproto/java` 中 `grpc.health.v1` 契约实现的标准健康检查服务（已注册服务返回 `SERVING`，未知服务返回 `NOT_FOUND`）；
- `GrpcServerLifecycle`：gRPC server 生命周期（启动/停止）与健康状态注册；
- 构建直接编译 `../../genproto/java`（已提交的生成代码），无 Docker / k8s / 云部署文件。
