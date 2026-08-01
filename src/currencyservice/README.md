# currencyservice (Node.js)

CurrencyService for the Online Boutique replica — implements
`hipstershop.CurrencyService/GetSupportedCurrencies` and `/Convert` (currency
conversion) plus the standard `grpc.health.v1.Health/Check` health service, per
`protos/currency.proto`.

Runtime dependencies: `@grpc/grpc-js` + `google-protobuf` (consumes the
committed `genproto/nodejs` stubs); CommonJS.

## Start

Workspace layout: the repo-root `package.json` (npm workspaces) holds the shared
`@grpc/grpc-js` / `google-protobuf` deps the `genproto/nodejs` stubs need, so run
`npm install` at the **repo root** once, then start from this directory:

```bash
# from repo root:
npm install

# from src/currencyservice:
npm start          # listens on 0.0.0.0:7000 (override: PORT=xxxx npm start)
```

## Test

```bash
# from repo root (runs all workspace tests) or from src/currencyservice:
npm test           # node --test (unit + boot/health smoke test)
```

## Health check

```bash
grpcurl -plaintext -d '{"service": "hipstershop.CurrencyService"}' localhost:7000 grpc.health.v1.Health/Check
# → {"status":"SERVING"}
```

## Behavior (aligned with upstream currencyservice)

- Exchange rates are static values relative to EUR (snapshot of the European
  Central Bank reference rates in `data/currency_conversion.json`); no external
  API, no middleware — logic runs entirely in-process.
- `GetSupportedCurrencies` returns the 33 supported ISO 4217 codes.
- `Convert` converts `from` (Money) into `to_code` via
  `from_currency → EUR → to_currency`, carrying decimals over the Money
  `units`/`nanos` representation (units/nanos floored in the result).
- Unknown currency codes return `INVALID_ARGUMENT` (gRPC code 3).
- `Health/Check` reports `SERVING` for any requested service name.
