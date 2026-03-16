FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/distance-service ./cmd/distance-service

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/distance-service /usr/local/bin/distance-service
EXPOSE 8082
CMD ["distance-service"]
