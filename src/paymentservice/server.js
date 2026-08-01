'use strict';

const grpc = require('@grpc/grpc-js');

const { PaymentServiceService } = require('../../genproto/nodejs/payment_grpc_pb.js');
const { ChargeResponse } = require('../../genproto/nodejs/payment_pb.js');
const { HealthService } = require('../../genproto/nodejs/grpc/health/v1/health_grpc_pb.js');
const { HealthCheckResponse } = require('../../genproto/nodejs/grpc/health/v1/health_pb.js');

const charge = require('./charge');
const logger = require('./logger');

const DEFAULT_PORT = '50051';

class PaymentServiceServer {
  constructor(port = process.env.PORT || DEFAULT_PORT) {
    this.port = port;
    this.server = new grpc.Server();
    this.server.addService(PaymentServiceService, {
      charge: PaymentServiceServer.chargeHandler,
    });
    this.server.addService(HealthService, {
      check: PaymentServiceServer.checkHandler,
    });
  }

  static chargeHandler(call, callback) {
    try {
      const { transaction_id: transactionId } = charge(call.request);
      const response = new ChargeResponse();
      response.setTransactionId(transactionId);
      callback(null, response);
    } catch (err) {
      logger.warn(`PaymentService#Charge rejected: ${err.message}`);
      callback(err);
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
          logger.info(`PaymentService gRPC server started on 0.0.0.0:${boundPort}`);
          resolve(boundPort);
        }
      );
    });
  }
}

module.exports = PaymentServiceServer;
