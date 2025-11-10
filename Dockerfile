# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    build-base \
    pkgconfig \
    gtk+3.0-dev \
    webkit2gtk-dev \
    nodejs \
    npm

# Install Wails
RUN go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2

WORKDIR /app
COPY . .

# Build frontend
RUN cd frontend && npm ci

# Build application
RUN wails build -platform linux/amd64

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache \
    gtk+3.0 \
    webkit2gtk \
    ca-certificates

WORKDIR /app

COPY --from=builder /app/build/bin/netcaptor /usr/local/bin/netcaptor

EXPOSE 8888 8080

CMD ["netcaptor"]
