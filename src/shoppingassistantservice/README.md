# shoppingassistantservice（Python）

本地复刻的**购物助手服务**：接收用户的聊天消息（`ShoppingAssistantService.GetCompletion`），
按规则分类意图（问候 / 商品推荐 / 商品详情 / 订单查询 / 闲聊）并给出回复。

> 上游 GoogleCloudPlatform/microservices-demo 的 shoppingassistantservice 是纯云 Flask 应用
> （Gemini + AlloyDB），无 gRPC 契约。本复刻按本地简化契约（见 `protos/shopping_assistant.proto`
> 注释）实现为规则 + 本地目录服务的 gRPC 助手，无任何云依赖。

- **契约**：`protos/shopping_assistant.proto` → `hipstershop.ShoppingAssistantService.GetCompletion`
- **消费**：ProductCatalogService（`product_catalog.proto`，默认 `localhost:3550`，可用环境变量
  `PRODUCT_CATALOG_SERVICE_ADDR` 覆盖）；目录不可达时服务优雅降级（返回兜底文案），不影响启动与健康检查
- **端口**：8082（`--port` 或环境变量 `PORT` 覆盖）
- **健康检查服务名**：`hipstershop.ShoppingAssistantService`

## 目录结构

```
src/shoppingassistantservice/
├── main.py        # 入口：启动 gRPC 服务（python main.py）
├── server.py      # ShoppingAssistantService 实现 + 标准 Health 服务注册
├── assistant.py   # 规则引擎：意图分类 + 回复生成（含多轮 follow-up 上下文）
├── catalog.py     # ProductCatalogService 的 gRPC 客户端（软失败）
├── paths.py       # 把仓库根 genproto/python 加入 sys.path
├── requirements.txt
└── tests/         # pytest 单测（假目录，无需外部服务）
```

## 启动

```bash
cd src/shoppingassistantservice
pip install -r requirements.txt
python main.py --port 8082
```

## 验证（gRPC health check）

```bash
grpcurl -plaintext -d '{"service": "hipstershop.ShoppingAssistantService"}' localhost:8082 grpc.health.v1.Health/Check
# => {"status":"SERVING"}
```

试一次对话：

```bash
grpcurl -plaintext -d '{"message": "recommend a gift"}' localhost:8082 hipstershop.ShoppingAssistantService/GetCompletion
grpcurl -plaintext -d '{"message": "tell me about the sunglasses"}' localhost:8082 hipstershop.ShoppingAssistantService/GetCompletion
```

## 测试

```bash
cd src/shoppingassistantservice
python -m pytest tests/ -v
```

单测覆盖：意图分类（问候/感谢/推荐/详情/订单/兜底）、推荐与详情回复内容、目录不可达时的
优雅降级、多轮 follow-up（"tell me more" 展开上次推荐）、`Money` 格式化，以及服务级冒烟测试
（Health Check 返回 SERVING、GetCompletion 往返、conversation_id 回显）。
