'use strict';

// Unit tests for the currency conversion core (no gRPC / network involved).

const test = require('node:test');
const assert = require('node:assert');

const currency = require('../currency');
const ratesData = require('../data/currency_conversion.json');

test('supportedCurrencies covers exactly the codes in the data file', () => {
  const codes = currency.supportedCurrencies();
  assert.deepStrictEqual([...codes].sort(), Object.keys(ratesData).sort());
  assert.ok(codes.includes('USD'));
  assert.ok(codes.includes('EUR'));
  assert.ok(codes.includes('JPY'));
});

test('rates are normalized to finite positive numbers', () => {
  for (const [code, rate] of currency.RATES) {
    assert.ok(Number.isFinite(rate), `${code} rate is finite`);
    assert.ok(rate > 0, `${code} rate is positive`);
  }
});

test('convert same currency is a no-op (approximately)', () => {
  const result = currency.convert({ currency_code: 'USD', units: 10, nanos: 990000000 }, 'USD');
  assert.strictEqual(result.currency_code, 'USD');
  // 10.99 USD -> 10.99 USD (small float drift allowed on nanos)
  assert.strictEqual(result.units, 10);
  assert.ok(Math.abs(result.nanos - 990000000) <= 1);
});

test('convert USD -> EUR scales by the USD/EUR rate', () => {
  // $10.99 at rate 1.1305 USD per EUR => ~9.7213 EUR
  const result = currency.convert({ currency_code: 'USD', units: 10, nanos: 990000000 }, 'EUR');
  assert.strictEqual(result.currency_code, 'EUR');
  assert.strictEqual(result.units, 9);
  assert.ok(result.nanos >= 700000000 && result.nanos <= 730000000, `nanos=${result.nanos}`);
});

test('convert EUR -> JPY scales by the EUR/JPY rate', () => {
  // 10 EUR at rate 126.40 JPY per EUR => 1264 JPY
  const result = currency.convert({ currency_code: 'EUR', units: 10, nanos: 0 }, 'JPY');
  assert.strictEqual(result.currency_code, 'JPY');
  assert.strictEqual(result.units, 1264);
  assert.strictEqual(result.nanos, 0);
});

test('convert round-trips USD -> EUR -> USD approximately', () => {
  const original = { currency_code: 'USD', units: 42, nanos: 500000000 };
  const inEur = currency.convert(original, 'EUR');
  const back = currency.convert(inEur, 'USD');
  assert.strictEqual(back.currency_code, 'USD');
  assert.strictEqual(back.units, original.units);
  assert.ok(Math.abs(back.nanos - original.nanos) <= 2, `nanos=${back.nanos}`);
});

test('convert rejects unknown from currency', () => {
  assert.throws(
    () => currency.convert({ currency_code: 'XXX', units: 1, nanos: 0 }, 'USD'),
    /unknown currency code: XXX/
  );
});

test('convert rejects unknown to currency', () => {
  assert.throws(
    () => currency.convert({ currency_code: 'USD', units: 1, nanos: 0 }, 'YYY'),
    /unknown currency code: YYY/
  );
});

test('carry folds fractional units into nanos', () => {
  // 0.5 units + 0 nanos -> 0 units + 500000000 nanos
  const result = currency.carry({ units: 0.5, nanos: 0 });
  assert.strictEqual(result.units, 0);
  assert.strictEqual(result.nanos, 500000000);
});

test('carry normalizes nanos overflow back into units', () => {
  // 1 unit + 1.5e9 nanos -> 2 units + 5e8 nanos
  const result = currency.carry({ units: 1, nanos: 1.5e9 });
  assert.strictEqual(result.units, 2);
  assert.strictEqual(result.nanos, 5e8);
});
