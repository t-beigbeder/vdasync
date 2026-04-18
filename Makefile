all:
	@echo "nothing is done when 'all' is done, try make help"

help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

vPATH := $(shell echo ${PATH})

.PHONY: test
test:	## go test the application
	go test ./...

.PHONY: test-again
test-again:	export QSTF_TEST_FULL = 1
test-again:	## go test the application again
	go test -count=1 ./...

.PHONY: build
build:	## go build all
	go build ./...

.PHONY: format
format:	## format go code
	gofmt -w .

.PHONY: grpc-code
grpc-code: export PATH=$(vPATH):$(HOME)/.local/bin:/home/dv-user/go/bin
grpc-code:	## generate grpc code from proto files
	protoc --go_out=. --go-grpc_out=. grpc/ope.proto
	protoc --go_out=. --go-grpc_out=. grpc/dssa.proto