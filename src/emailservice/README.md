# emailservice（Python）

复刻 GoogleCloudPlatform/microservices-demo（Online Boutique）的 emailservice。
消费 `protos/email.proto`（`hipstershop.EmailService.SendOrderConfirmation`）
与 `protos/grpc/health/v1/health.proto`（标准健康检查），实现代码使用
`genproto/python` 生成桩。

本地范围：渲染订单确认邮件（Jinja2 模板 `templates/confirmation.html`）并模拟发送
（默认仅打印日志）；配置 `SMTP_HOST` 后可经真实 SMTP 中继发送。

## 本地启动

```bash
pip install -r requirements.txt
python email_server.py                 # 默认端口 8080（可用 --port 或 $PORT 覆盖）
```

启动后日志输出 `emailservice listening on port 8080`。

## 验证

1. 启动服务（见上）。
2. 单元测试：

   ```bash
   pip install pytest
   pytest -v
   ```

3. gRPC 健康检查（期望 `SERVING`）：

   ```bash
   grpcurl -plaintext -d '{"service": "hipstershop.EmailService"}' localhost:8080 grpc.health.v1.Health/Check
   ```

4. 发送一条订单确认（模拟模式，服务端日志可见收件人与订单号）：

   ```bash
   python email_client.py --email alice@example.com --order-id 12345
   ```

## 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | gRPC 监听端口 |
| `SMTP_HOST` | 未设置 | 设置后走真实 SMTP 发送，否则本地模拟 |
| `SMTP_PORT` | `587` | SMTP 端口 |
| `SMTP_USER` / `SMTP_PASSWORD` | 未设置 | SMTP 认证 |
| `FROM_ADDRESS` | `no-reply@example.com` | 发件人地址 |
| `EMAILSERVICE_DEBUG` | 未设置 | 设置后输出 DEBUG 日志（含邮件 HTML 正文） |

## 目录结构

```
src/emailservice/
├── email_server.py        # gRPC 服务端（EmailService + Health）
├── email_client.py        # 命令行测试客户端
├── templates/             # 邮件模板（Jinja2）
│   └── confirmation.html
├── tests/                 # pytest 单元测试（冒烟级）
│   └── test_emailservice.py
└── requirements.txt       # 运行时依赖
```
