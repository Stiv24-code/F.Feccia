/**
 * Logger condizionale — logga solo in development, silenzioso in produzione.
 */
const isDev = import.meta.env.DEV;

type LogFn = (...args: unknown[]) => void;

interface Logger {
  error: LogFn;
  warn: LogFn;
  info: LogFn;
  debug: LogFn;
}

export const logger: Logger = {
  error: (...args) => { if (isDev) console.error('[LoginBusiness]', ...args); },
  warn: (...args) => { if (isDev) console.warn('[LoginBusiness]', ...args); },
  info: (...args) => { if (isDev) console.info('[LoginBusiness]', ...args); },
  debug: (...args) => { if (isDev) console.debug('[LoginBusiness]', ...args); },
};
