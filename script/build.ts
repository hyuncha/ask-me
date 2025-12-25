import { execSync } from 'child_process';
import { build } from 'esbuild';
import path from 'path';
import fs from 'fs';

const frontendDir = path.join(process.cwd(), 'frontend');
const backendDir = path.join(process.cwd(), 'backend');

async function main() {
  console.log('Building frontend...');
  
  try {
    execSync('npm install', { 
      cwd: frontendDir, 
      stdio: 'inherit' 
    });
    
    execSync('npm run build', { 
      cwd: frontendDir, 
      stdio: 'inherit' 
    });
    
    console.log('Frontend build completed successfully!');
  } catch (error) {
    console.error('Frontend build failed:', error);
    process.exit(1);
  }

  console.log('Building Go backend (static binary)...');
  
  try {
    execSync('go build -buildvcs=false -ldflags="-s -w" -o bin/api ./cmd/api', {
      cwd: backendDir,
      stdio: 'inherit',
      env: { 
        ...process.env, 
        HOME: '/tmp',
        CGO_ENABLED: '0',
        GOOS: 'linux',
        GOARCH: 'amd64'
      }
    });
    console.log('Go backend build completed successfully!');
  } catch (error) {
    console.error('Go backend build failed:', error);
    process.exit(1);
  }

  console.log('Building server bundle...');
  
  try {
    await build({
      entryPoints: ['server/index.ts'],
      bundle: true,
      platform: 'node',
      target: 'node20',
      format: 'cjs',
      outfile: 'dist/index.cjs',
      external: ['pg-native'],
    });
    console.log('Server bundle completed successfully!');
  } catch (error) {
    console.error('Server bundle failed:', error);
    process.exit(1);
  }

  console.log('Build completed!');
}

main();
