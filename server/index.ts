import { spawn, ChildProcess } from 'child_process';
import { parse } from 'url';
import path from 'path';
import fs from 'fs';

const parseDbUrl = () => {
  const dbUrl = process.env.DATABASE_URL;
  if (!dbUrl) return;

  try {
    const url = new URL(dbUrl);
    process.env.DB_USER = url.username;
    process.env.DB_PASSWORD = url.password;
    process.env.DB_HOST = url.hostname;
    process.env.DB_PORT = url.port || '5432';
    process.env.DB_NAME = url.pathname.slice(1).split('?')[0];
    process.env.DB_SSL_MODE = process.env.NODE_ENV === 'production' ? 'require' : 'disable';
  } catch (e) {
    console.error('Failed to parse DATABASE_URL:', e);
  }
};

const startGoBackend = (): ChildProcess => {
  parseDbUrl();

  const backendDir = path.resolve(process.cwd(), 'backend');
  const binaryPath = path.join(backendDir, 'bin', 'api');

  console.log('Starting Cleaners AI Go Backend...');
  console.log(`Working directory: ${backendDir}`);
  console.log(`Binary path: ${binaryPath}`);
  console.log(`DB_HOST: ${process.env.DB_HOST}`);
  console.log(`DB_PORT: ${process.env.DB_PORT}`);
  console.log(`DB_NAME: ${process.env.DB_NAME}`);
  console.log(`DB_SSL_MODE: ${process.env.DB_SSL_MODE}`);
  console.log(`SERVER_PORT: ${process.env.SERVER_PORT || '5000'}`);

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
