#!/bin/bash

echo "Starting Cleaners AI on Replit..."

# Start PostgreSQL
if [ ! -d "$HOME/postgresql" ]; then
  mkdir -p $HOME/postgresql/data
  initdb -D $HOME/postgresql/data
fi

# Configure PostgreSQL
cat > $HOME/postgresql/data/postgresql.conf <<EOF
listen_addresses = 'localhost'
port = 5432
max_connections = 20
shared_buffers = 128MB
EOF

cat > $HOME/postgresql/data/pg_hba.conf <<EOF
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
EOF

# Start PostgreSQL server
pg_ctl -D $HOME/postgresql/data -l $HOME/postgresql/logfile start

# Wait for PostgreSQL to start
sleep 3

# Create database if not exists
psql -U $USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'cleaners_ai'" | grep -q 1 || psql -U $USER -d postgres -c "CREATE DATABASE cleaners_ai"

# Run init.sql if it exists
if [ -f "./backend/scripts/init.sql" ]; then
  psql -U $USER -d cleaners_ai -f ./backend/scripts/init.sql
fi

# Start Redis
redis-server --daemonize yes --port 6379

# Wait for Redis to start
sleep 2

# Install backend dependencies
cd backend
go mod download
cd ..

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build frontend
cd frontend
REACT_APP_API_URL=$REPL_SLUG.$REPL_OWNER.repl.co npm run build
cd ..

# Start backend server
cd backend
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=$USER \
DB_PASSWORD="" \
DB_NAME=cleaners_ai \
DB_SSL_MODE=disable \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
SERVER_PORT=8080 \
ENVIRONMENT=production \
PINECONE_API_KEY=$PINECONE_API_KEY \
PINECONE_ENV=$PINECONE_ENV \
PINECONE_INDEX_NAME=$PINECONE_INDEX_NAME \
LLM_API_KEY=$LLM_API_KEY \
LLM_MODEL=$LLM_MODEL \
go run cmd/api/main.go &

BACKEND_PID=$!
cd ..

# Start frontend server
cd frontend
npx serve -s build -l 3000 &
FRONTEND_PID=$!
cd ..

echo "Backend started with PID: $BACKEND_PID"
echo "Frontend started with PID: $FRONTEND_PID"
echo "Cleaners AI is running!"
echo "Backend API: http://localhost:8080"
echo "Frontend: http://localhost:3000"

# Wait for both processes
wait
