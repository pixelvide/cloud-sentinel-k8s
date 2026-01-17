# User-defined build args
ARG FRONTEND_PORT=3000

# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@latest --activate
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY ui/ .
RUN pnpm run build

# Stage 2: Build Backend
FROM public.ecr.aws/docker/library/golang:1.25-alpine AS backend-builder

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY version.txt .
RUN VERSION=$(cat version.txt | tr -d '\n\r') && \
    go build -ldflags="-X main.Version=$VERSION" -o server main.go

# Stage 3: Final Image
FROM public.ecr.aws/docker/library/alpine:latest

# Install necessary runtime dependencies
RUN apk add --no-cache curl glab aws-cli ca-certificates git openssh libc6-compat

# Shim host path for glab and aws
RUN mkdir -p /opt/homebrew/bin /usr/local/bin && \
    ln -sf /usr/bin/glab /opt/homebrew/bin/glab && \
    ln -sf /usr/bin/aws /usr/local/bin/aws

WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/dist ./static

EXPOSE 8080
CMD ["./server"]
