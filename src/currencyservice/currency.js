'use strict';

// Currency conversion core for currencyservice.
//
// Behavior is aligned with the upstream Online Boutique currencyservice
// (GoogleCloudPlatform/microservices-demo/src/currencyservice/server.js):
// - rates are static values relative to EUR (European Central Bank reference
//   rates snapshot, see ./data/currency_conversion.json); no external API
// - conversion goes from_currency -> EUR -> to_currency with decimal/fractional
//   carrying over the Money units/nanos representation.

const ratesData = require('./data/currency_conversion.json');

// Stored rates are strings like "1.1305"; normalize to numbers once.
const RATES = new Map(
  Object.entries(ratesData).map(([code, rate]) => [code, Number(rate)])
);

const NANOS_PER_UNIT = 1e9;

// Sorted list of the 3-letter ISO 4217 codes we can convert between.
function supportedCurrencies() {
  return Array.from(RATES.keys());
}

function getRate(code) {
  return RATES.get(code);
}

// Handles decimal/fractional carrying over the Money representation
// (upstream _carry): folds the fractional part of `units` into `nanos` and
// re-normalizes so nanos stays within [0, 1e9) and units stays integral.
function carry(amount) {
  amount.nanos += (amount.units % 1) * NANOS_PER_UNIT;
  amount.units = Math.floor(amount.units) + Math.floor(amount.nanos / NANOS_PER_UNIT);
  amount.nanos = amount.nanos % NANOS_PER_UNIT;
  return amount;
}

// Converts `from` ({currency_code, units, nanos}) into `toCode`.
// Returns {currency_code, units, nanos} (units/nanos floored, nanos in
// [0, 999999999]); throws RangeError on unknown currency codes.
function convert(from, toCode) {
  const fromRate = getRate(from.currency_code);
  if (fromRate === undefined) {
    throw new RangeError(`unknown currency code: ${from.currency_code}`);
  }
  const toRate = getRate(toCode);
  if (toRate === undefined) {
    throw new RangeError(`unknown currency code: ${toCode}`);
  }

  // from_currency --> EUR
  const euros = carry({
    units: from.units / fromRate,
    nanos: from.nanos / fromRate,
  });
  euros.nanos = Math.round(euros.nanos);

  // EUR --> to_currency
  const result = carry({
    units: euros.units * toRate,
    nanos: euros.nanos * toRate,
  });

  result.units = Math.floor(result.units);
  result.nanos = Math.floor(result.nanos);
  result.currency_code = toCode;
  return result;
}

module.exports = {
  RATES,
  NANOS_PER_UNIT,
  supportedCurrencies,
  getRate,
  carry,
  convert,
};
