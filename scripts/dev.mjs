import { spawn } from 'child_process';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = resolve(__dirname, '..');

const processes = [];

function cleanup() {
  console.log('\nShutting down...');
  processes.forEach(p => {
    if (p && !p.killed) {
      p.kill('SIGTERM');
    }
  });
  process.exit(0);
}

process.on('SIGINT', cleanup);
process.on('SIGTERM', cleanup);

const backendEnv = { ...process.env };
if (process.env.DATABASE_URL) {
  const url = process.env.DATABASE_URL;
  const clean = url.replace('postgresql://', '');
  const [creds, rest] = clean.split('@');
  const [user, pass] = creds.split(':');
  const [hostPort, dbPart] = rest.split('/');
  const [host, port] = hostPort.split(':');
  const dbName = dbPart.split('?')[0];
  
  backendEnv.DB_USER = user;
  backendEnv.DB_PASSWORD = pass;
  backendEnv.DB_HOST = host;
  backendEnv.DB_PORT = port;
  backendEnv.DB_NAME = dbName;
  backendEnv.DB_SSL_MODE = 'require';
}

console.log('Starting Go backend on port 8080...');
const backend = spawn('./bin/api', [], {
  cwd: resolve(rootDir, 'backend'),
  env: backendEnv,
  stdio: ['ignore', 'pipe', 'pipe']
});
processes.push(backend);

backend.stdout.on('data', (data) => {
  console.log(`[backend] ${data.toString().trim()}`);
});

backend.stderr.on('data', (data) => {
  console.error(`[backend] ${data.toString().trim()}`);
});

backend.on('error', (err) => {
  console.error(`[backend] Failed to start: ${err.message}`);
});

backend.on('exit', (code) => {
  console.log(`[backend] Exited with code ${code}`);
});

await new Promise(resolve => setTimeout(resolve, 2000));

console.log('Starting React frontend on port 5000...');
const frontend = spawn('npm', ['start'], {
  cwd: resolve(rootDir, 'frontend'),
  env: { ...process.env, PORT: '5000', BROWSER: 'none' },
  stdio: ['ignore', 'pipe', 'pipe'],
  shell: true
});
processes.push(frontend);

frontend.stdout.on('data', (data) => {
  console.log(`[frontend] ${data.toString().trim()}`);
});

frontend.stderr.on('data', (data) => {
  const msg = data.toString().trim();
  if (msg) console.error(`[frontend] ${msg}`);
});

frontend.on('error', (err) => {
  console.error(`[frontend] Failed to start: ${err.message}`);
});

frontend.on('exit', (code) => {
  console.log(`[frontend] Exited with code ${code}`);
});

console.log('Development server started. Go API on :8080, React on :5000');
