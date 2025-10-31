# ---- Build stage ----
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Cache deps first for faster rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary (adjust ./main.go if your entrypoint is different)
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/bblog ./main.go

# ---- Runtime stage ----
FROM alpine:3.20
WORKDIR /app

# Create non-root user to run the app
RUN adduser -D -H -u 10001 appuser

# Copy the built artifact from the builder
COPY --from=builder /out/bblog /app/bblog

# (Optional) persistent dir for any writable paths your app needs
RUN mkdir -p /app/uploads && chown -R appuser:appuser /app

ENV GIN_MODE=release \ 
    PORT=8080
EXPOSE 8080

USER appuser
ENTRYPOINT ["/app/bblog"]
