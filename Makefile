APP_NAME   := api
APP_DIR := ./cmd/api
DIST_DIR = ./dist

setup:
	go mod download

test:
	go test -parallel=6 -failfast -cover ./...
	
build: ## Build go binary
	go build -o ${DIST_DIR}/${APP_NAME} ${APP_DIR}

buildcron: 
	go build -o ${DIST_DIR}/cron ./cmd/cron/main.go

local: 
	docker-compose up