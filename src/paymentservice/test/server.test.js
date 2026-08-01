'use strict';

// Smoke test: boots the real gRPC server on an ephemeral port and verifies
// both the business RPC (hipstershop.PaymentService/Charge) and the standard
// health RPC (grpc.health.v1.Health/Check) over the generated clients.

const test = require('node:test');
const assert = require('node:assert');
const grpc = require('@grpc/grpc-js');

const PaymentServiceServer = require('../server');
const { PaymentServiceClient } = require('../../../genproto/nodejs/payment_grpc_pb.js');
const { ChargeRequest } = require('../../../genproto/nodejs/payment_pb.js');
const { Money, CreditCardInfo } = require('../../../genproto/nodejs/demo_pb.js');
const { HealthClient } = require('../../../genproto/nodejs/grpc/health/v1/health_grpc_pb.js');
const {
  HealthCheckRequest,
  HealthCheckResponse,
} = require('../../../genproto/nodejs/grpc/health/v1/health_pb.js');

test('smoke: server boots and answers Charge + Health/Check', async (t) => {
  const server = new PaymentServiceServer('0');
  const boundPort = await server.listen();
  const addr = `localhost:${boundPort}`;

  t.after(() => {
    server.server.forceShutdown();
  });

  // --- hipstershop.PaymentService/Charge ---
  await new Promise((resolve, reject) => {
    const client = new PaymentServiceClient(addr, grpc.credentials.createInsecure());

    const creditCard = new CreditCardInfo();
    creditCard.setCreditCardNumber('4111111111111111');
    creditCard.setCreditCardCvv(123);
    creditCard.setCreditCardExpirationYear(new Date().getFullYear() + 1);
    creditCard.setCreditCardExpirationMonth(12);

    const amount = new Money();
    amount.setCurrencyCode('USD');
    amount.setUnits(10);
    amount.setNanos(99);

    const req = new ChargeRequest();
    req.setAmount(amount);
    req.setCreditCard(creditCard);

    client.charge(req, (err, res) => {
      client.close();
      if (err) {
        reject(err);
        return;
      }
      try {
        assert.ok(res.getTransactionId());
        assert.match(res.getTransactionId(), /^[0-9a-f]{8}-/);
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  });

  // --- grpc.health.v1.Health/Check ---
  await new Promise((resolve, reject) => {
    const client = new HealthClient(addr, grpc.credentials.createInsecure());
    const req = new HealthCheckRequest();
    req.setService('hipstershop.PaymentService');

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
