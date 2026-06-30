# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/cost-optimizer ./cmd

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /app/cost-optimizer ./cost-optimizer
COPY --from=builder /src/web ./web

EXPOSE 8080
ENTRYPOINT ["./cost-optimizer"]
