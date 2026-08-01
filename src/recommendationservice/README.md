# recommendationservice（Python）

复刻 GoogleCloudPlatform/microservices-demo（Online Boutique）的
`recommendationservice`，gRPC 服务端，契约见 `protos/recommendation.proto`。

- 服务名（健康检查）：`hipstershop.RecommendationService`
- 默认端口：`8080`
- 消费：`productcatalogservice`（`hipstershop.ProductCatalogService.ListProducts`）
- 逻辑对齐上游：拉取全量商品目录 → 剔除购物车内已有商品 → 随机返回最多 5 个
  推荐商品 ID。

## 目录

```
src/recommendationservice/
├── recommendation_server.py   # gRPC 服务端（入口，python 直接运行）
├── logger.py                  # JSON 日志（仅标准库）
├── conftest.py                # pytest 路径引导（genproto/python + 服务目录）
├── requirements.txt           # 运行与测试依赖
└── tests/
    └── test_recommendation_server.py  # 冒烟测试（单测 + 进程内 E2E + health check）
```

## 本地启动

```bash
# 1) 安装依赖（Python 3.11+）
pip install -r src/recommendationservice/requirements.txt

# 2) 启动（默认 PRODUCT_CATALOG_SERVICE_ADDR=localhost:3550，PORT=8080）
python src/recommendationservice/recommendation_server.py

# 或指定 productcatalogservice 地址与端口
PRODUCT_CATALOG_SERVICE_ADDR=localhost:3550 PORT=8080 \
  python src/recommendationservice/recommendation_server.py
```

环境变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | gRPC 监听端口 |
| `PRODUCT_CATALOG_SERVICE_ADDR` | `localhost:3550` | productcatalogservice 地址（`host:port`） |

> gRPC channel 为惰性连接：未配置真实的 productcatalogservice 时服务仍可启动，
> 健康检查照常通过；调用 `ListRecommendations` 时才需要目录服务可达。

## 单元测试

```bash
cd src/recommendationservice
python -m pytest tests/ -v
```

覆盖：推荐逻辑（剔除购物车商品、上限 5 条、空目录）、进程内 gRPC E2E
（真实 `ListRecommendations` 往返）、`grpc.health.v1.Health/Check` 返回 `SERVING`。

## gRPC health check

服务启动后：

```bash
grpcurl -plaintext -d '{"service": "hipstershop.RecommendationService"}' \
  localhost:8080 grpc.health.v1.Health/Check
```

预期返回：`{"status": "SERVING"}`。
