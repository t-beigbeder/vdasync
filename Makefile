all:
	@echo "nothing is done when 'all' is done, try make help"

help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

vPATH := $(shell echo ${PATH})

.PHONY: test
test:	## go test the application
	go test -v ./...

.PHONY: test-again
test-again:	export QSTF_TEST_FULL = 1
test-again:	## go test the application again
	go test -v -count=1 ./...

.PHONY: build-test
build-test:	## go build test cmd
	go build -o bin/test cmd/test/main.go
	go build -o bin/testgrpc cmd/testgrpc/main.go

.PHONY: run-test
run-test:	## go build and run test cmd
run-test: build-test
	@echo skip bin/test -is-fatal
	bin/testgrpc

build: ## go build all
build: build-test

.PHONY: format
format:	## format go code
	gofmt -w .

.PHONY: grpc-code
grpc-code: export PATH=$(vPATH):$(HOME)/.local/bin:/home/dv-user/go/bin
grpc-code:	## generate grpc code from proto files
	protoc --go_out=. --go-grpc_out=. grpc/ope.proto
	protoc --go_out=. --go-grpc_out=. grpc/dssa.proto