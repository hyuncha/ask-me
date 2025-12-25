#!/bin/bash

parse_db_url() {
    local url="$DATABASE_URL"
    
    local clean_url="${url#postgresql://}"
    local creds="${clean_url%%@*}"
    local rest="${clean_url#*@}"
    
    export DB_USER="${creds%%:*}"
    export DB_PASSWORD="${creds#*:}"
    
    local host_port="${rest%%/*}"
    local db_part="${rest#*/}"
    
    export DB_HOST="${host_port%%:*}"
    export DB_PORT="${host_port#*:}"
    export DB_NAME="${db_part%%\?*}"
    export DB_SSL_MODE="require"
}

if [ -n "$DATABASE_URL" ]; then
    parse_db_url
fi

echo "Starting Cleaners AI Go Backend..."
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_NAME: $DB_NAME"
echo "SERVER_PORT: $SERVER_PORT"

cd /home/runner/workspace/backend
exec ./bin/api
