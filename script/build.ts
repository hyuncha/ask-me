import { execSync } from "child_process";
import { build } from "esbuild";
import path from "path";

const frontendDir = path.join(process.cwd(), "frontend");
const backendDir = path.join(process.cwd(), "backend");

const isReplitDeploy = process.env.REPLIT_DEPLOYMENT === "1"
  || process.env.REPLIT_DEPLOYMENT === "true"
  || process.env.REPLIT_DEPLOY === "true"
  || process.env.DEPLOYMENT === "true";

async function main() {
  console.log("[build] Building frontend...");
  
  try {
    execSync("npm install --legacy-peer-deps", { 
      cwd: frontendDir, 
      stdio: "inherit" 
    });
    
    execSync("SKIP_PREFLIGHT_CHECK=true npm run build", { 
      cwd: frontendDir, 
      stdio: "inherit" 
    });
    
    console.log("[build] Frontend build completed successfully!");
  } catch (error) {
    console.error("[build] Frontend build failed:", error);
    process.exit(1);
  }

  if (isReplitDeploy) {
    console.log("[build] Replit Deploy build env: skipping Go build");
  } else {
    console.log("[build] Building Go backend binary...");
    
    try {
      execSync("go build -buildvcs=false -ldflags=\"-s -w\" -o bin/api ./cmd/api", {
        cwd: backendDir,
        stdio: "inherit",
        env: { 
          ...process.env, 
          HOME: "/tmp",
          CGO_ENABLED: "0",
          GOOS: "linux",
          GOARCH: "amd64"
        }
      });
      console.log("[build] Go backend build completed successfully!");
    } catch (error) {
      console.error("[build] Go backend build failed:", error);
      process.exit(1);
    }
  }

  console.log("[build] Building server bundle...");
  
  try {
    await build({
      entryPoints: ["server/index.ts"],
      bundle: true,
      platform: "node",
      target: "node20",
      format: "cjs",
      outfile: "dist/index.cjs",
      external: ["pg-native"],
    });
    console.log("[build] Server bundle completed successfully!");
  } catch (error) {
    console.error("[build] Server bundle failed:", error);
    process.exit(1);
  }

  console.log("[build] Build completed!");
}

main();
