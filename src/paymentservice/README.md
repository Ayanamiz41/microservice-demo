# paymentservice (Node.js)

PaymentService for the Online Boutique replica — implements
`hipstershop.PaymentService/Charge` (credit-card payment processing) and the
standard `grpc.health.v1.Health/Check` health service, per `protos/payment.proto`.

Runtime dependencies: `@grpc/grpc-js` + `google-protobuf` (consumes the
committed `genproto/nodejs` stubs); CommonJS.

## Start

Workspace layout: the repo-root `package.json` (npm workspaces) holds the shared
`@grpc/grpc-js` / `google-protobuf` deps the `genproto/nodejs` stubs need, so run
`npm install` at the **repo root** once, then start from this directory:

```bash
# from repo root:
npm install

# from src/paymentservice:
npm start          # listens on 0.0.0.0:50051 (override: PORT=xxxx npm start)
```

## Test

```bash
# from repo root (runs all workspace tests) or from src/paymentservice:
npm test           # node --test (unit + boot/health smoke test)
```

## Health check

```bash
grpcurl -plaintext -d '{"service": "hipstershop.PaymentService"}' localhost:50051 grpc.health.v1.Health/Check
# → {"status":"SERVING"}
```

## Behavior (aligned with upstream paymentservice)

- `Charge` validates the credit card number, accepts only VISA / MasterCard,
  rejects expired cards, and returns a random `transaction_id`.
- Invalid / unaccepted / expired cards return `INVALID_ARGUMENT` (gRPC code 3).
- `Health/Check` reports `SERVING` for any requested service name.
