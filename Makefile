.PHONY: keys deps build run-auth run-weather run-distance run-gateway run-frontend dev docker-up docker-down

# ─── Key Generation ───────────────────────────────────────────────────────────
keys:
	@echo "Generating RSA key pair..."
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "✓ Keys generated in ./keys/"

# ─── Go Dependencies ──────────────────────────────────────────────────────────
deps:
	go mod tidy

# ─── Build All Services ───────────────────────────────────────────────────────
build: deps
	go build ./cmd/auth-service/...
	go build ./cmd/weather-service/...
	go build ./cmd/distance-service/...
	go build ./cmd/api-gateway/...
	@echo "✓ All services built"

# ─── Run Individual Services ──────────────────────────────────────────────────
run-auth:
	PRIVATE_KEY_PATH=./keys/private.pem PUBLIC_KEY_PATH=./keys/public.pem PORT=8083 go run ./cmd/auth-service

run-weather:
	PUBLIC_KEY_PATH=./keys/public.pem PORT=8081 go run ./cmd/weather-service

run-distance:
	PUBLIC_KEY_PATH=./keys/public.pem PORT=8082 go run ./cmd/distance-service

run-gateway:
	PORT=8080 \
	AUTH_SERVICE_URL=http://localhost:8083 \
	WEATHER_SERVICE_URL=http://localhost:8081 \
	DISTANCE_SERVICE_URL=http://localhost:8082 \
	go run ./cmd/api-gateway

run-frontend:
	cd frontend && npm run dev

# ─── Dev (all services in background + frontend) ─────────────────────────────
dev: keys
	@echo "Starting all services..."
	@mkdir -p .pids
	PUBLIC_KEY_PATH=./keys/public.pem PORT=8081 go run ./cmd/weather-service &
	PUBLIC_KEY_PATH=./keys/public.pem PORT=8082 go run ./cmd/distance-service &
	PRIVATE_KEY_PATH=./keys/private.pem PUBLIC_KEY_PATH=./keys/public.pem PORT=8083 go run ./cmd/auth-service &
	sleep 2
	PORT=8080 AUTH_SERVICE_URL=http://localhost:8083 WEATHER_SERVICE_URL=http://localhost:8081 DISTANCE_SERVICE_URL=http://localhost:8082 go run ./cmd/api-gateway &
	cd frontend && npm install && npm run dev

# ─── Docker ──────────────────────────────────────────────────────────────────
docker-up: keys
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
