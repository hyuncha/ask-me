import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';

/**
 * Parse database configuration from environment variables
 * Priority: DATABASE_URL > PGHOST/DB_HOST env vars
 * NO localhost fallback - missing config is a fatal error
 */
const parseDbConfig = (): void => {
  // Priority 1: DATABASE_URL (supports both TCP and Unix socket)
  const dbUrl = process.env.DATABASE_URL;
  if (dbUrl) {
    try {
      const url = new URL(dbUrl);
      process.env.DB_USER = url.username;
      process.env.DB_PASSWORD = url.password;
      process.env.DB_NAME = url.pathname.slice(1).split('?')[0];

      // Parse query parameters for sslmode and socket host
      const params = new URLSearchParams(url.search);
      const socketHost = params.get('host');

      const sslModeParam = params.get('sslmode');

      if (socketHost && socketHost.startsWith('/')) {
        // Unix socket: postgres://user:pass@/dbname?host=/cloudsql/project:region:instance
        process.env.DB_HOST = socketHost;
        process.env.DB_PORT = '5432'; // Ignored for Unix socket but set for consistency
        // Unix socket is already secure, disable is OK
        process.env.DB_SSL_MODE = sslModeParam || 'disable';
        console.log(`[DB Config] socket=${socketHost} db=${process.env.DB_NAME} sslmode=${process.env.DB_SSL_MODE}`);
      } else if (url.hostname) {
        // TCP connection
        process.env.DB_HOST = url.hostname;
        process.env.DB_PORT = url.port || '5432';
        // TCP requires SSL
        process.env.DB_SSL_MODE = (sslModeParam && sslModeParam !== 'disable') ? sslModeParam : 'require';
        console.log(`[DB Config] host=${url.hostname} port=${process.env.DB_PORT} db=${process.env.DB_NAME} sslmode=${process.env.DB_SSL_MODE}`);
      } else {
        console.error('[DB Config] DATABASE_URL must specify host or socket path');
        process.exit(1);
      }
      return;
    } catch (e) {
      console.error('[DB Config] Failed to parse DATABASE_URL:', e);
      process.exit(1);
    }
  }

  // Priority 2: Individual environment variables (PGHOST or DB_HOST)
  const host = process.env.DB_HOST || process.env.PGHOST;
  if (!host) {
    console.error('[DB Config] ERROR: DATABASE_URL or DB_HOST/PGHOST is required - localhost fallback is disabled');
    process.exit(1);
  }

  // Validate: reject localhost explicitly
  if (host === 'localhost' || host === '127.0.0.1') {
    console.error('[DB Config] ERROR: localhost database connection is not allowed - use Cloud SQL socket or remote host');
    process.exit(1);
  }

  process.env.DB_HOST = host;
  process.env.DB_PORT = process.env.DB_PORT || process.env.PGPORT || '5432';
  process.env.DB_USER = process.env.DB_USER || process.env.PGUSER;
  process.env.DB_PASSWORD = process.env.DB_PASSWORD || process.env.PGPASSWORD || '';
  process.env.DB_NAME = process.env.DB_NAME || process.env.PGDATABASE;

  if (!process.env.DB_USER) {
    console.error('[DB Config] ERROR: DB_USER/PGUSER is required');
    process.exit(1);
  }

  if (!process.env.DB_NAME) {
    console.error('[DB Config] ERROR: DB_NAME/PGDATABASE is required');
    process.exit(1);
  }

  // SSL: Unix socket is already secure (disable OK), TCP requires SSL
  const sslMode = process.env.DB_SSL_MODE || process.env.PGSSLMODE;
  if (host.startsWith('/')) {
    // Unix socket - disable is OK
    process.env.DB_SSL_MODE = sslMode || 'disable';
  } else {
    // TCP - require SSL
    process.env.DB_SSL_MODE = (sslMode && sslMode !== 'disable') ? sslMode : 'require';
  }

  // Log connection target (no secrets)
  if (host.startsWith('/')) {
    console.log(`[DB Config] socket=${host} db=${process.env.DB_NAME} sslmode=${process.env.DB_SSL_MODE}`);
  } else {
    console.log(`[DB Config] host=${host} port=${process.env.DB_PORT} db=${process.env.DB_NAME} sslmode=${process.env.DB_SSL_MODE}`);
  }
};

const startGoBackend = (): ChildProcess => {
  parseDbConfig();

  const cwd = process.cwd();
  const backendDir = path.resolve(cwd, 'backend');
  const binaryPath = path.join(backendDir, 'bin', 'api');
  const staticDir = path.resolve(cwd, 'frontend', 'build');

  // Set STATIC_DIR for Go backend
  if (!process.env.STATIC_DIR) {
    process.env.STATIC_DIR = staticDir;
  }

  // Map OPENAI_API_KEY to LLM_API_KEY if not set
  if (!process.env.LLM_API_KEY && process.env.OPENAI_API_KEY) {
    process.env.LLM_API_KEY = process.env.OPENAI_API_KEY;
  }

  console.log('[Startup] Starting Cleaners AI Go Backend...');
  console.log(`[Startup] Binary: ${binaryPath}`);
  console.log(`[Startup] Static dir: ${process.env.STATIC_DIR} (exists: ${fs.existsSync(process.env.STATIC_DIR)})`);
  console.log(`[Startup] SERVER_PORT: ${process.env.SERVER_PORT || '5000'}`);
  console.log(`[Startup] LLM_API_KEY set: ${!!process.env.LLM_API_KEY}`);

  if (!fs.existsSync(binaryPath)) {
    console.error(`Go binary not found at: ${binaryPath}`);
    console.error('Please run "npm run build" first to compile the Go backend.');
    process.exit(1);
  }

  const backend = spawn(binaryPath, [], {
    cwd: backendDir,
    env: process.env as NodeJS.ProcessEnv,
    stdio: 'inherit',
  });

  backend.on('error', (err) => {
    console.error('Go backend failed to start:', err.message);
    process.exit(1);
  });

  backend.on('exit', (code) => {
    console.log(`Go backend exited with code ${code}`);
    process.exit(code || 0);
  });

  return backend;
};

const handleSignals = (backend: ChildProcess) => {
  const cleanup = () => {
    console.log('\nShutting down...');
    backend.kill('SIGTERM');
  };

  process.on('SIGINT', cleanup);
  process.on('SIGTERM', cleanup);
};

const backend = startGoBackend();
handleSignals(backend);
