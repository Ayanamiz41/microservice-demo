'use strict';

// Minimal leveled logger (stdout). Dependency-free stand-in for pino.
// Levels: debug < info < warn < error; runtime level from LOG_LEVEL env.

const LEVELS = { debug: 10, info: 20, warn: 30, error: 40 };
const currentLevel = LEVELS[process.env.LOG_LEVEL] || LEVELS.info;

function log(level, args) {
  if (LEVELS[level] < currentLevel) {
    return;
  }
  const message = args
    .map((arg) => (typeof arg === 'string' ? arg : JSON.stringify(arg)))
    .join(' ');
  process.stdout.write(
    `${new Date().toISOString()} ${level.toUpperCase()} [currencyservice] ${message}\n`
  );
}

module.exports = {
  debug: (...args) => log('debug', args),
  info: (...args) => log('info', args),
  warn: (...args) => log('warn', args),
  error: (...args) => log('error', args),
};
