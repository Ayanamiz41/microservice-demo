'use strict';

const PaymentServiceServer = require('./server');
const logger = require('./logger');

const server = new PaymentServiceServer();

server.listen().catch((err) => {
  logger.error(`Failed to start PaymentService: ${err.message}`);
  process.exit(1);
});
