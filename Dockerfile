# ---------- (1) Backend build stage ----------
FROM golang:1.22-alpine AS go-builder
WORKDIR /app

# Go 모듈 파일 먼저 복사 (캐시 이점)
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# 백엔드 소스 복사 후 빌드
COPY backend ./backend
RUN cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

# ---------- (2) Frontend build stage ----------
FROM node:20-alpine AS fe-builder
WORKDIR /app

# frontend 디렉토리의 의존성 먼저
COPY frontend/package*.json ./frontend/
RUN cd frontend && npm ci --legacy-peer-deps

# frontend 소스 복사 후 빌드
COPY frontend ./frontend
RUN cd frontend && SKIP_PREFLIGHT_CHECK=true npm run build

# ---------- (3) Server build stage ----------
FROM node:20-alpine AS server-builder
WORKDIR /app

# 루트 의존성 설치
COPY package*.json ./
RUN npm ci

# 서버 코드 복사 및 번들링
COPY server ./server
COPY tsconfig.json ./
RUN npx esbuild server/index.ts --bundle --platform=node --target=node20 --format=cjs --outfile=dist/index.cjs --external:pg-native

# ---------- (4) Runtime stage ----------
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

# 서버 번들 복사
COPY --from=server-builder /app/dist ./dist
COPY --from=server-builder /app/package*.json ./

# 프론트엔드 빌드 결과물 복사
COPY --from=fe-builder /app/frontend/build ./frontend/build

# Go 바이너리 복사
RUN mkdir -p backend/bin
COPY --from=go-builder /out/api ./backend/bin/api

# 런타임 의존성만 설치
RUN npm ci --omit=dev

EXPOSE 8080
CMD ["node", "dist/index.cjs"]
