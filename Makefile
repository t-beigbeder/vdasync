all:
	@echo "nothing is done when 'all' is done, try make help"

help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

vPATH := $(shell echo ${PATH})
CERTS_PATH := $(shell if [ -z "${CERTS_PATH}" ] ; then echo /local/tmp/certs ; else echo ${CERTS_PATH} ; fi)
CERT_SERVER := $(shell if [ -z "${CERT_SERVER}" ] ; then hostname ; else echo ${CERT_SERVER} ; fi)

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
	go build -o bin/testcerts cmd/testcerts/main.go
	go build -o bin/vdaserver cmd/vdaserver/main.go

.PHONY: certs
certs:	## generate test certificates
	@mkdir -p $(CERTS_PATH)
	go build -o bin/testcerts cmd/testcerts/main.go
	bin/testcerts -ca $(CERTS_PATH)/sca-cert.pem -cakey $(CERTS_PATH)/sca-key.pem -cn Server-CA
	@echo openssl x509 -in $(CERTS_PATH)/sca-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -cn Client-CA
	@echo openssl x509 -in $(CERTS_PATH)/cca-cert.pem -text -noout
	bin/testcerts -cert $(CERTS_PATH)/self-cert.pem -key $(CERTS_PATH)/self-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/self-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/sca-cert.pem -cakey $(CERTS_PATH)/sca-key.pem -hosts localhost,$(CERT_SERVER) -cert $(CERTS_PATH)/localhost-cert.pem -key $(CERTS_PATH)/localhost-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/localhost-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -hosts localhost -cert $(CERTS_PATH)/plugin-cert.pem -key $(CERTS_PATH)/plugin-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/plugin-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -cert $(CERTS_PATH)/client-cert.pem -key $(CERTS_PATH)/client-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/client-cert.pem -text -noout

.PHONY: format
format:	## format go code
	gofmt -w .

.PHONY: grpc-code
grpc-code: export PATH=$(vPATH):$(HOME)/.local/bin:/home/dv-user/go/bin
grpc-code:	## generate grpc code from proto files
	protoc --go_out=. --go-grpc_out=. grpc/ope.proto
	protoc --go_out=. --go-grpc_out=. grpc/dssa.proto