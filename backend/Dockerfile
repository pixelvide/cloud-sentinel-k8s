FROM public.ecr.aws/docker/library/golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY version.txt .
RUN export VERSION=$(cat version.txt) && \
    go build -ldflags="-X main.Version=$VERSION" -o server main.go

FROM public.ecr.aws/docker/library/alpine:latest

# Install necessary runtime dependencies
RUN apk add --no-cache curl glab aws-cli ca-certificates git openssh libc6-compat

# Shim host path for glab and aws
# Kubeconfig often points to absolute paths like /opt/homebrew/bin/glab or /usr/local/bin/aws
RUN mkdir -p /opt/homebrew/bin /usr/local/bin && \
    ln -sf /usr/bin/glab /opt/homebrew/bin/glab && \
    ln -sf /usr/bin/aws /usr/local/bin/aws

WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
