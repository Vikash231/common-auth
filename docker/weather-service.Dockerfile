FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/weather-service ./cmd/weather-service

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/weather-service /usr/local/bin/weather-service
EXPOSE 8081
CMD ["weather-service"]
