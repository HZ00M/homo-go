GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
#	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	Git_Bash="D:/Git/bin/bash.exe"
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

.PHONY: service
service:
	@for file in $(API_PROTO_FILES); do \
		# 取proto文件的基本名，比如 api/helloworld/v1/demo.proto -> demo.go \
		basefile=$$(basename $$file .proto).go; \
		targetfile=internal/service/$$basefile; \
		if [ ! -f $$targetfile ]; then \
			echo "Generating server for $$file"; \
			kratos proto server $$file -t internal/service; \
		else \
			echo "Skip $$file, target $$targetfile exists"; \
		fi \
	done
.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api;
	make config;
	make generate;

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
# wire 生成规则
.PHONY: wire
wire:
	@echo ">> Generating wire dependency injection..."
	@go install github.com/google/wire/cmd/wire@latest
	@wire ./..

.PHONY: mock
# 查找所有 *_grpc.pb.go 文件
MOCK_SOURCE_FILES := $(shell find api -name "*_grpc.pb.go")
MOCK_TARGET_DIR := mocks

mock:
	@mkdir -p $(MOCK_TARGET_DIR)
	@echo "Generating mocks for all gRPC services..."
	@for file in $(MOCK_SOURCE_FILES); do \
		proto_file=$$(basename $$file | sed 's/_grpc.pb.go/.proto/'); \
		mock_file_name=$$(echo $$proto_file | sed 's/\.proto$$/_mock.go/'); \
		mock_out_path=$(MOCK_TARGET_DIR)/$$mock_file_name; \
		echo "Generating mock from $$file to $$mock_out_path"; \
		mockgen -source=$$file -destination=$$mock_out_path -package=mocks; \
	done


