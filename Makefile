prepare:
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@if [ ! -f .air.toml ]; then \
        echo "Creating .air.toml file..."; \
        air init; \
    else \
        echo ".air.toml already exists, skipping..."; \
    fi
	swag init
	go mod tidy
	go mod download

r:
	docker-compose down -v
	npx kill-port 8080
	docker-compose up
