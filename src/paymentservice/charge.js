'use strict';

// Charge logic for PaymentService, ported from upstream
// GoogleCloudPlatform/microservices-demo src/paymentservice/charge.js.
// Works against the generated hipstershop.ChargeRequest message.

const cardValidator = require('simple-card-validator');
const { v4: uuidv4 } = require('uuid');
const grpc = require('@grpc/grpc-js');

const logger = require('./logger');

class CreditCardError extends Error {
  constructor(message) {
    super(message);
    this.code = grpc.status.INVALID_ARGUMENT; // invalid argument
  }
}

class InvalidCreditCard extends CreditCardError {
  constructor() {
    super('Credit card info is invalid');
  }
}

class UnacceptedCreditCard extends CreditCardError {
  constructor(cardType) {
    super(
      `Sorry, we cannot process ${cardType} credit cards. Only VISA or MasterCard is accepted.`
    );
  }
}

class ExpiredCreditCard extends CreditCardError {
  constructor(number, month, year) {
    super(`Your credit card (ending ${number.substr(-4)}) expired on ${month}/${year}`);
  }
}

/**
 * Verifies the credit card number and (pretend) charges the card.
 *
 * @param {ChargeRequest} request - generated hipstershop.ChargeRequest message
 * @return {{transaction_id: string}}
 */
function charge(request) {
  const amount = request.getAmount();
  const creditCard = request.getCreditCard();
  const cardNumber = creditCard.getCreditCardNumber();
  const cardInfo = cardValidator(cardNumber);
  const { card_type: cardType, valid } = cardInfo.getCardDetails();

  if (!valid) {
    throw new InvalidCreditCard();
  }

  // Only VISA and MasterCard are accepted; other card types (AMEX, dinersclub)
  // throw UnacceptedCreditCard.
  if (!(cardType === 'visa' || cardType === 'mastercard')) {
    throw new UnacceptedCreditCard(cardType);
  }

  // Expiration must be in the future.
  const currentMonth = new Date().getMonth() + 1;
  const currentYear = new Date().getFullYear();
  const year = creditCard.getCreditCardExpirationYear();
  const month = creditCard.getCreditCardExpirationMonth();
  if (currentYear * 12 + currentMonth > year * 12 + month) {
    throw new ExpiredCreditCard(cardNumber.replace('-', ''), month, year);
  }

  logger.info(
    `Transaction processed: ${cardType} ending ${cardNumber.substr(-4)} ` +
      `Amount: ${amount.getCurrencyCode()}${amount.getUnits()}.${amount.getNanos()}`
  );

  return { transaction_id: uuidv4() };
}

module.exports = charge;
