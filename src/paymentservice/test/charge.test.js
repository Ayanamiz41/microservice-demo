'use strict';

const test = require('node:test');
const assert = require('node:assert');
const grpc = require('@grpc/grpc-js');

const { ChargeRequest } = require('../../../genproto/nodejs/payment_pb.js');
const { Money, CreditCardInfo } = require('../../../genproto/nodejs/demo_pb.js');

const charge = require('../charge');

const FUTURE_YEAR = new Date().getFullYear() + 1;

function makeRequest({ number, year, month, units = 10, nanos = 99 }) {
  const creditCard = new CreditCardInfo();
  creditCard.setCreditCardNumber(number);
  creditCard.setCreditCardCvv(123);
  creditCard.setCreditCardExpirationYear(year);
  creditCard.setCreditCardExpirationMonth(month);

  const amount = new Money();
  amount.setCurrencyCode('USD');
  amount.setUnits(units);
  amount.setNanos(nanos);

  const req = new ChargeRequest();
  req.setAmount(amount);
  req.setCreditCard(creditCard);
  return req;
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

test('charge succeeds for a valid VISA card and returns a transaction id', () => {
  const res = charge(makeRequest({ number: '4111111111111111', year: FUTURE_YEAR, month: 12 }));
  assert.ok(res.transaction_id);
  assert.match(res.transaction_id, UUID_RE);
});

test('charge succeeds for a valid MasterCard', () => {
  const res = charge(makeRequest({ number: '5555555555554444', year: FUTURE_YEAR, month: 1 }));
  assert.ok(res.transaction_id);
  assert.match(res.transaction_id, UUID_RE);
});

test('charge rejects an invalid card number with INVALID_ARGUMENT', () => {
  assert.throws(
    () => charge(makeRequest({ number: '1234', year: FUTURE_YEAR, month: 12 })),
    (err) =>
      err.message.includes('invalid') && err.code === grpc.status.INVALID_ARGUMENT
  );
});

test('charge rejects unsupported card types (AMEX) with INVALID_ARGUMENT', () => {
  assert.throws(
    () => charge(makeRequest({ number: '378282246310005', year: FUTURE_YEAR, month: 12 })),
    (err) =>
      err.message.includes('VISA or MasterCard') && err.code === grpc.status.INVALID_ARGUMENT
  );
});

test('charge rejects an expired card with INVALID_ARGUMENT', () => {
  assert.throws(
    () => charge(makeRequest({ number: '4111111111111111', year: 2020, month: 1 })),
    (err) =>
      err.message.includes('expired') && err.code === grpc.status.INVALID_ARGUMENT
  );
});

test('charge rejects a card expiring this month (not yet expired) - boundary ok', () => {
  const now = new Date();
  const res = charge(
    makeRequest({ number: '4111111111111111', year: now.getFullYear(), month: now.getMonth() + 1 })
  );
  assert.ok(res.transaction_id);
});
