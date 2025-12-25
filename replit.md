# Cleaners AI

## Overview

Cleaners AI is an AI-powered laundry knowledge service that provides users with expert advice on laundry-related topics. The application features a chat interface where users can ask questions about stain removal, fabric care, laundry techniques, and more. Users can also upload images (laundry labels, stains) for more accurate AI-powered responses.

The system includes:
- A consumer-facing chat interface for laundry advice
- An admin panel for managing knowledge base content
- User authentication via Google OAuth
- Multi-language support (Korean and English)

## Recent Changes

- **2025-12-21**: Integrated Go backend from GitHub repository (dewdew-list/cleaners-ai)
  - Set up Node.js wrapper in `server/index.ts` to spawn Go binary
  - Configured database connection with SSL disabled for Replit Postgres
  - Added SPA static file serving to Go router
  - Created all database tables (users, conversations, messages, knowledge_items, etc.)

## User Preferences

Preferred communication style: Simple, everyday language.

## System Architecture

### Backend Architecture (Go)

The backend is a **Go-based API server** using clean architecture:
- **Location**: `/backend` directory
- **Binary**: `backend/bin/api` (compiled with `go build -buildvcs=false`)
- **Entry point**: `backend/cmd/api/main.go`
- **Architecture layers**:
  - `domain/entity` - Core domain models
  - `application/service` - Business logic services
  - `infrastructure/persistence` - Database repositories
  - `interface/http/handler` - HTTP handlers and router
- **Wrapper**: `server/index.ts` - Node.js script that parses DATABASE_URL and spawns Go binary

### Frontend Architecture (React)

Located in `/frontend` directory:
- React 18 with Create React App
- React Router for navigation
- Axios for API communication
- TypeScript for type safety
- Proxy configured to Go backend at localhost:8080

### Data Layer

- **Database**: PostgreSQL (Replit built-in)
- **Tables**: users, conversations, messages, knowledge_items, query_logs, subscription_plans, user_subscriptions
- **Connection**: Go backend parses DB_* environment variables from DATABASE_URL

### Authentication

- Google OAuth integration (requires GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET)
- JWT-based session management
- Token stored in localStorage

### API Endpoints

- `GET /health` - Health check
- `POST /api/chat/message` - Send chat message
- `GET /api/chat/conversations` - List conversations
- `GET /api/chat/history/:id` - Get conversation history
- `POST /api/upload` - Upload file
- `POST /api/knowledge` - Create knowledge item
- `GET /api/knowledge/search` - Search knowledge base
- `POST /api/extract-text` - Extract text from documents
- `GET /auth/google` - Initiate Google OAuth
- `GET /auth/google/callback` - OAuth callback
- `POST /auth/refresh` - Refresh JWT token
- `GET /auth/me` - Get current user

## Environment Variables

### Required for full functionality:
- `DATABASE_URL` - PostgreSQL connection string (auto-provided by Replit)
- `SERVER_PORT` - API server port (set to 5000)
- `LLM_API_KEY` - OpenAI API key for chat functionality
- `GOOGLE_CLIENT_ID` - Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth client secret

### Optional:
- `PINECONE_API_KEY` - For vector search (RAG features)
- `PINECONE_ENV` - Pinecone environment
- `PINECONE_INDEX_NAME` - Pinecone index name
- `STRIPE_SECRET_KEY` - For payment processing
- `STRIPE_PUBLISHABLE_KEY` - Stripe public key

### Current configuration:
- `DB_SSL_MODE=disable` - Required for Replit Postgres (no SSL)
- `ENVIRONMENT=development`
- `JWT_SECRET` - Set for development

## Development

### Starting the server:
```bash
npm run dev
```
This runs `server/index.ts` which spawns the Go backend on port 5000.

### Rebuilding Go backend:
```bash
cd backend && HOME=/tmp go build -buildvcs=false -o bin/api ./cmd/api
```

### Testing:
```bash
curl http://localhost:5000/health  # Should return "OK"
```

## Deployment Notes

- For production deployments with managed Postgres, set `DB_SSL_MODE=require`
- Frontend build files should be placed in `frontend/build/` for SPA serving
- Go binary is pre-compiled and checked into repository