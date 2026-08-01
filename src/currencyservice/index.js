'use strict';

const CurrencyServiceServer = require('./server');
const logger = require('./logger');

const server = new CurrencyServiceServer();

server.listen().catch((err) => {
  logger.error(`Failed to start CurrencyService: ${err.message}`);
  process.exit(1);
});
