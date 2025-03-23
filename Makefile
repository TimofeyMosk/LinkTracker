COVERAGE_FILE ?= coverage.out
BOT_BINARY=./bin/bot
SCRAPPER_BINARY=./bin/scrapper

# Сборка обоих приложений
.PHONY: build
build: build_bot build_scrapper

# Запуск обоих приложений
.PHONY: run
run: build
	@echo "Запуск bot и scrapper в фоновом режиме"
	@$(BOT_BINARY) & \
	$(SCRAPPER_BINARY) & \
	echo "Приложения запущены. Используйте 'make stop' для остановки."

# Остановка всех запущенных приложений
.PHONY: stop
stop:
	@echo "Остановка всех запущенных приложений"
	@pkill -f $(BOT_BINARY) || true
	@pkill -f $(SCRAPPER_BINARY) || true
	@echo "Приложения остановлены."

# Сборка bot
.PHONY: build_bot
build_bot:
	@echo "Сборка bot..."
	@mkdir -p ./bin
	@go build -o $(BOT_BINARY) ./cmd/bot

# Сборка scrapper
.PHONY: build_scrapper
build_scrapper:
	@echo "Сборка scrapper..."
	@mkdir -p ./bin
	@go build -o $(SCRAPPER_BINARY) ./cmd/scrapper

# Запуск bot
.PHONY: run_bot
run_bot:
	@echo "Запуск bot..."
	@$(BOT_BINARY)

# Запуск scrapper
.PHONY: run_scrapper
run_scrapper:
	@echo "Запуск scrapper..."
	@$(SCRAPPER_BINARY)


## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='./...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

.PHONY: lint
lint: lint-golang #lint-proto

.PHONY: lint-golang
lint-golang:
	@if ! command -v 'golangci-lint' &> /dev/null; then \
  		echo "Please install golangci-lint!"; exit 1; \
  	fi;
	@golangci-lint -v run --fix ./...

#.PHONY: lint-proto
#lint-proto:
#	@if ! command -v 'easyp' &> /dev/null; then \
#  		echo "Please install easyp!"; exit 1; \
#	fi;
#	@easyp lint

.PHONY: generate
generate: generate_openapi  #generate_proto

#.PHONY: generate_proto
#generate_proto:
#	@if ! command -v 'easyp' &> /dev/null; then \
#		echo "Please install easyp!"; exit 1; \
#	fi;
#	@easyp generate

.PHONY: generate_openapi
generate_openapi:
	@if ! command -v 'oapi-codegen' &> /dev/null; then \
		echo "Please install oapi-codegen!"; exit 1; \
	fi;
	@mkdir -p internal/api/openapi/v1
	@oapi-codegen -package botdto \
		-generate types \
		api/openapi/bot-api.yaml > internal/infrastructure/dto/dto_bot/bot-api.gen.go
	@oapi-codegen -package scrapperdto \
    		-generate types \
    		api/openapi/scrapper-api.yaml > internal/infrastructure/dto/dto_scrapper/scrapper-api.gen.go

.PHONY: clean
clean:
	@rm -rf./bin