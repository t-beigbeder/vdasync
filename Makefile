all:
	@echo "nothing is done when 'all' is done, try make help"

help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

vPATH := $(shell echo ${PATH})

.PHONY: test
test: export GO_TEST_LOG_LEVEL = ERROR
test:	## go test the application
	go test -v ./...

.PHONY: test-verbose
test-verbose:	## go test the application
	go test -v ./...

.PHONY: test-this
test-this:	## go test the application
	go test -v -run TestMakeTestFilesTree github.com/t-beigbeder/vdasync/internal/common

.PHONY: test-again
test-again:	export OTVL_TEST_FULL = 1
test-again: export GO_TEST_LOG_LEVEL = ERROR
test-again:	## go test the application again
	go test -v -count=1 ./...

.PHONY: test-again-verbose
test-again-verbose:	export OTVL_TEST_FULL = 1
test-again-verbose:	## go test the application again
	go test -v -count=1 ./...

.PHONY: build
build:	## go build commands
	go build -o bin/localFiles cmd/plugins/localfiles/main.go
	go build -o bin/vdasync cmd/vdasync/main.go

.PHONY: format
format:	## format go code
	gofmt -w .

.PHONY: grpc-code
grpc-code: export PATH=$(vPATH):$(HOME)/.local/bin:/home/dv-user/go/bin
grpc-code:	## generate grpc code from proto files
	protoc --go_out=. --go-grpc_out=. grpc/ope.proto
	protoc --go_out=. --go-grpc_out=. grpc/dssa.proto