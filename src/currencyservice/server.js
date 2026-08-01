'use strict';

const grpc = require('@grpc/grpc-js');

const { CurrencyServiceService } = require('../../genproto/nodejs/currency_grpc_pb.js');
const { GetSupportedCurrenciesResponse } = require('../../genproto/nodejs/currency_pb.js');
const { Money } = require('../../genproto/nodejs/demo_pb.js');
const { HealthService } = require('../../genproto/nodejs/grpc/health/v1/health_grpc_pb.js');
const { HealthCheckResponse } = require('../../genproto/nodejs/grpc/health/v1/health_pb.js');

const currency = require('./currency');
const logger = require('./logger');

const DEFAULT_PORT = '7000';

class CurrencyServiceServer {
  constructor(port = process.env.PORT || DEFAULT_PORT) {
    this.port = port;
    this.server = new grpc.Server();
    this.server.addService(CurrencyServiceService, {
      getSupportedCurrencies: CurrencyServiceServer.getSupportedCurrenciesHandler,
      convert: CurrencyServiceServer.convertHandler,
    });
    this.server.addService(HealthService, {
      check: CurrencyServiceServer.checkHandler,
    });
  }

  static getSupportedCurrenciesHandler(call, callback) {
    try {
      const response = new GetSupportedCurrenciesResponse();
      response.setCurrencyCodesList(currency.supportedCurrencies());
      logger.info(`GetSupportedCurrencies -> ${response.getCurrencyCodesList().length} codes`);
      callback(null, response);
    } catch (err) {
      logger.warn(`GetSupportedCurrencies failed: ${err.message}`);
      callback(err);
    }
  }

  static convertHandler(call, callback) {
    try {
      const from = call.request.getFrom();
      if (!from) {
        throw new RangeError('missing from amount');
      }
      const result = currency.convert(
        {
          currency_code: from.getCurrencyCode(),
          units: from.getUnits(),
          nanos: from.getNanos(),
        },
        call.request.getToCode()
      );
      const money = new Money();
      money.setCurrencyCode(result.currency_code);
      money.setUnits(result.units);
      money.setNanos(result.nanos);
      logger.info(
        `Convert ${from.getUnits()}.${from.getNanos()} ${from.getCurrencyCode()} -> ` +
          `${result.units}.${result.nanos} ${result.currency_code}`
      );
      callback(null, money);
    } catch (err) {
      logger.warn(`Convert rejected: ${err.message}`);
      callback({ code: grpc.status.INVALID_ARGUMENT, message: err.message });
    }
  }

  static checkHandler(call, callback) {
    const response = new HealthCheckResponse();
    response.setStatus(HealthCheckResponse.ServingStatus.SERVING);
    callback(null, response);
  }

  listen() {
    return new Promise((resolve, reject) => {
      this.server.bindAsync(
        `0.0.0.0:${this.port}`,
        grpc.ServerCredentials.createInsecure(),
        (err, boundPort) => {
          if (err) {
            reject(err);
            return;
          }
          // grpc-js >=1.x starts serving automatically once bound; explicit
          // start() is deprecated and can be omitted.
          logger.info(`CurrencyService gRPC server started on 0.0.0.0:${boundPort}`);
          resolve(boundPort);
        }
      );
    });
  }
}

module.exports = CurrencyServiceServer;
