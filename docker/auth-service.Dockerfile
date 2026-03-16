FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/auth-service ./cmd/auth-service

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/auth-service /usr/local/bin/auth-service
EXPOSE 8083
CMD ["auth-service"]
