# Stage 1: Build Frontend (Next.js)
FROM public.ecr.aws/docker/library/node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# Install dependencies (use cache mount to speed up)
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

# Copy source and build
COPY frontend/ ./
# Disable telemetry during build
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Stage 2: Build Backend (Go)
FROM public.ecr.aws/docker/library/golang:1.25-alpine AS backend-builder
WORKDIR /app/backend

# Install build dependencies
RUN apk add --no-cache git

# Download Go modules
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source, version file and build
COPY version.txt .
COPY backend/ ./
RUN export VERSION=$(grep -oP '(?<=version=).*' version.txt) && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.Version=$VERSION" -o server main.go

# Stage 3: Final Production Image
FROM public.ecr.aws/docker/library/alpine:latest

# Install runtime dependencies (inc. tools for k8s interaction)
RUN apk add --no-cache ca-certificates curl glab aws-cli git openssh libc6-compat

# Shim host paths for glab and aws to match potential user kubeconfig paths
RUN mkdir -p /opt/homebrew/bin /usr/local/bin && \
    ln -sf /usr/bin/glab /opt/homebrew/bin/glab && \
    ln -sf /usr/bin/aws /usr/local/bin/aws

WORKDIR /app

# Copy Backend Binary
COPY --from=backend-builder /app/backend/server ./server

# Copy Frontend Assets (Next.js standalone build)
# Note: Next.js 'standalone' output includes a minimal server.js
# For a truly unified binary, we usually serve the static assets via the Go server.
# However, Next.js 'standalone' mode is designed to run its own node process.
#
# STRATEGY: 
# 1. We will run BOTH processes using a lightweight process manager (supervisord) OR
# 2. If the goal is "one container, one entrypoint", we can serve the 'out' (static export) via Go.
#    BUT, the current frontend uses 'standalone' which implies SSR features might be used.
#    Given the user request "combine node and golang output", we will copy the standalone build.
#    Since we need to run the Go API primarily, we will assume the User wants to serve everything
#    from the Single Docker Image. 
#
#    Current Implementation:
#    - We need Node.js in the final image to run the Next.js standalone server.
#    - We will use a script to start both, OR just install nodejs in the final image.

RUN apk add --no-cache nodejs npm

# Copy Frontend Standalone Build
COPY --from=frontend-builder /app/frontend/.next/standalone ./frontend
COPY --from=frontend-builder /app/frontend/.next/static ./frontend/.next/static
COPY --from=frontend-builder /app/frontend/public ./frontend/public

# Setup Environment Strings
ENV NODE_ENV=production
ENV PORT=3000
# Backend listens on 8080, Frontend on 3000

# Create a start script to run both
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Starting Backend on :8080..."' >> /app/start.sh && \
    echo './server &' >> /app/start.sh && \
    echo 'echo "Starting Frontend on :3000..."' >> /app/start.sh && \
    echo 'cd frontend && node server.js' >> /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 3000 8080

CMD ["/app/start.sh"]
