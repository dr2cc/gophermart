include .env
export

export LOCAL_BIN:=$(CURDIR)/bin
export PATH:=$(LOCAL_BIN):$(PATH)

run: ## just 'make' to run app
	go run ./cmd/gophermart/main.go
.PHONY: run

help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

drop: ## сбросить БД и запустить приложение
	go run ./cmd/gophermart/main.go -drop
.PHONY: drop

test: ## run tests
	go test -v ./...
.PHONY: test

# swag: ## run swag
# 	swag init -g cmd/main.go
# .PHONY: swag

# build: ## run build
# 	docker-compose build todo-app

# up: ## run up
# 	docker-compose up todo-app

# migrate: ## run migrate
# 	migrate -path ./schema -database 'postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable' up

# так объявляем, что это комманды, а не переменные:
# .PHONY: help run drop
#
# Как работать с Makefile в VSCode https://victorz.ru/202402043262
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html