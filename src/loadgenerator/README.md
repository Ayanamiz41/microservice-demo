# loadgenerator（Python / Locust）

负载生成服务：Locust 压测脚本，模拟真实用户购物流程（浏览首页 → 切换币种 →
浏览商品 → 加购 → 查看购物车 → 结算下单），对齐上游
GoogleCloudPlatform/microservices-demo 的 `src/loadgenerator`。

本服务为**纯客户端**：不暴露任何 gRPC 服务端（无服务契约、无健康检查端口），
只通过 **frontend 的 HTTP REST 接口**施加负载；frontend 内部再以 gRPC 扇出到
各后端服务（productcatalog / cart / checkout / currency / ad / recommendation /
shipping / payment / email）。消费路径见 `src/README.md` 服务矩阵。

## 本地运行

```bash
# 1) 安装依赖（locust + faker，版本与上游一致）
pip install -r requirements.txt

# 2) 启动压测（默认无头模式；frontend 默认 http://localhost:8080）
#    直接使用 locust 命令（在服务目录下，自动发现 locustfile.py）：
cd src/loadgenerator
locust --host http://localhost:8080 --headless -u 10 -r 1

# 或等价的 python 入口（仓库根目录，Git Bash / Linux / macOS）：
PYTHONPATH=src python -m loadgenerator --host http://localhost:8080 --headless -u 10 -r 1

# Windows PowerShell 等价 python 入口：
#   $env:PYTHONPATH = "src"
#   python -m loadgenerator --host http://localhost:8080 --headless -u 10 -r 1
```

去掉 `--headless` 时启动 Locust Web UI（默认 `http://localhost:8089`），
可交互式设置用户数并实时查看指标。完整参数与 `locust --help` 一致。

> 说明：本任务验收不要求在本地真正跑压测（需要先启动 frontend 及后端服务），
> 只要求脚本可启动、单测全绿。

## 单元测试

```bash
python -m pytest tests -q
```

覆盖：locustfile 可正常导入且任务图与上游一致、加购/结算表单载荷合法、
genproto/python 生成客户端可导入（gRPC 消费链路）、`python -m loadgenerator`
入口接线正确。

## 文件说明

| 文件 | 说明 |
| --- | --- |
| `locustfile.py` | Locust 压测脚本（任务权重/场景与上游一致） |
| `__main__.py` | `python -m loadgenerator` 入口（透传 Locust CLI） |
| `requirements.in` / `requirements.txt` | 依赖（locust==2.43.0, faker==40.1.0） |
| `tests/test_locustfile.py` | 冒烟级单测 |
