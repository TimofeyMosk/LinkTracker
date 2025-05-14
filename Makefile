COVERAGE_FILE ?= coverage.out

.PHONY: build
build:
	@docker compose up -d --build

.PHONY: run
run:
	@docker compose up -d
	@echo "Приложения запущены. Используйте 'make stop' для остановки."

.PHONY: stop
stop:
	@docker compose down
	@echo "Приложения остановлены."




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
	@docker compose down -v