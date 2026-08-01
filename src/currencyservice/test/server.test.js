'use strict';

// Smoke test: boots the real gRPC server on an ephemeral port and verifies
// both business RPCs (hipstershop.CurrencyService/GetSupportedCurrencies,
// /Convert) and the standard health RPC (grpc.health.v1.Health/Check) over
// the generated clients.

const test = require('node:test');
const assert = require('node:assert');
const grpc = require('@grpc/grpc-js');

const CurrencyServiceServer = require('../server');
const { CurrencyServiceClient } = require('../../../genproto/nodejs/currency_grpc_pb.js');
const { CurrencyConversionRequest } = require('../../../genproto/nodejs/currency_pb.js');
const { Empty, Money } = require('../../../genproto/nodejs/demo_pb.js');
const { HealthClient } = require('../../../genproto/nodejs/grpc/health/v1/health_grpc_pb.js');
const {
  HealthCheckRequest,
  HealthCheckResponse,
} = require('../../../genproto/nodejs/grpc/health/v1/health_pb.js');

function startServer(t) {
  const server = new CurrencyServiceServer('0');
  return server.listen().then((boundPort) => {
    t.after(() => {
      server.server.forceShutdown();
    });
    return `localhost:${boundPort}`;
  });
}

test('smoke: GetSupportedCurrencies returns the currency codes', async (t) => {
  const addr = await startServer(t);

  await new Promise((resolve, reject) => {
    const client = new CurrencyServiceClient(addr, grpc.credentials.createInsecure());
    client.getSupportedCurrencies(new Empty(), (err, res) => {
      client.close();
      if (err) {
        reject(err);
        return;
      }
      try {
        const codes = res.getCurrencyCodesList();
        assert.ok(codes.length >= 30, `got ${codes.length} codes`);
        assert.ok(codes.includes('USD'));
        assert.ok(codes.includes('EUR'));
        assert.ok(codes.includes('JPY'));
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  });
});

test('smoke: Convert converts USD to EUR over the wire', async (t) => {
  const addr = await startServer(t);

  await new Promise((resolve, reject) => {
    const client = new CurrencyServiceClient(addr, grpc.credentials.createInsecure());

    const from = new Money();
    from.setCurrencyCode('USD');
    from.setUnits(10);
    from.setNanos(990000000);

    const req = new CurrencyConversionRequest();
    req.setFrom(from);
    req.setToCode('EUR');

    client.convert(req, (err, res) => {
      client.close();
      if (err) {
        reject(err);
        return;
      }
      try {
        assert.strictEqual(res.getCurrencyCode(), 'EUR');
        assert.strictEqual(res.getUnits(), 9);
        assert.ok(res.getNanos() >= 700000000 && res.getNanos() <= 730000000);
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  });
});

test('smoke: Convert rejects an unknown currency code', async (t) => {
  const addr = await startServer(t);

  await new Promise((resolve, reject) => {
    const client = new CurrencyServiceClient(addr, grpc.credentials.createInsecure());

    const from = new Money();
    from.setCurrencyCode('USD');
    from.setUnits(1);
    from.setNanos(0);

    const req = new CurrencyConversionRequest();
    req.setFrom(from);
    req.setToCode('XXX');

    client.convert(req, (err) => {
      client.close();
      if (!err) {
        reject(new Error('expected INVALID_ARGUMENT for unknown currency code'));
        return;
      }
      try {
        assert.strictEqual(err.code, grpc.status.INVALID_ARGUMENT);
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  });
});

test('smoke: Health/Check reports SERVING', async (t) => {
  const addr = await startServer(t);

  await new Promise((resolve, reject) => {
    const client = new HealthClient(addr, grpc.credentials.createInsecure());
    const req = new HealthCheckRequest();
    req.setService('hipstershop.CurrencyService');

    client.check(req, (err, res) => {
      client.close();
      if (err) {
        reject(err);
        return;
      }
      try {
        assert.strictEqual(
          res.getStatus(),
          HealthCheckResponse.ServingStatus.SERVING
        );
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  });
});
