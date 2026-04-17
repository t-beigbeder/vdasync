all:
	@echo "nothing is done when 'all' is done, try make help"
help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
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
.PHONY: build
format:	## format go code
	gofmt -w .
