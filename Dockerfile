# ---------- (1) Frontend build stage ----------
FROM node:20-alpine AS fe-builder
WORKDIR /app

# 의존성 먼저
COPY package*.json ./
RUN npm ci

# 소스 복사 후 빌드
COPY . .
RUN npm run build

# ---------- (2) Backend build stage ----------
FROM golang:1.22-alpine AS go-builder
WORKDIR /app

# Go 모듈 파일 먼저 복사 (캐시 이점)
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# 백엔드 소스 복사 후 빌드
COPY backend ./backend
RUN cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

# ---------- (3) Runtime stage ----------
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

# Node 런타임(서버 런처) 파일들 복사
# server/index.ts를 dist로 빌드해 쓰는 구조라면 dist만 복사
COPY --from=fe-builder /app/dist ./dist
COPY --from=fe-builder /app/package*.json ./

# (선택) 서버 런처 코드가 dist 외에 필요하면 추가 COPY
COPY server ./server
COPY scripts ./scripts

# Go 바이너리 복사
RUN mkdir -p backend/bin
COPY --from=go-builder /out/api ./backend/bin/api

# 런타임 의존성만 설치 (필요 시)
RUN npm ci --omit=dev

EXPOSE 8080
# ✅ 여기서 server/index가 backend/bin/api를 실행하는 구조면 그대로 OK
CMD ["node", "server/index.js"]
