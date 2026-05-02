# otvl_dtacsy

Versatile data access and synchronization tool

## Status

Pre-alpha status:

- basic API local and remote
- plugin setup with local implementation for testing
- walker structure for synchronization implementation
- rather good test coverage

## Design

### Dssa: data storage system abstraction

Features

- local files
- grpc server to remote local files or any plugin
- plugin through same grpc API
  - grpc avoids having large and bloat binary
  - s3
  - sftp client

API, cli-tool, both through configuration file for enhanced operability and better security

API

- list
- get
- put
- delete
- stat, setStat

### Fast synchronization with parallelism

Dssa walk requests pushed on worker queue

## Dev

### Go

Random useful commands

    go clean -modcache
    go get github.com/goccy/go-yaml
    go build -o bin/manager cmd/main.go

### code-server golang extension

[go extension](https://github.com/golang/vscode-go/wiki/)
[run/debug](https://github.com/golang/vscode-go/wiki/debugging#launchjson-attributes)

    go install -v github.com/go-delve/delve/cmd/dlv@latest

### Protobuf

https://protobuf.dev/installation/

    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go get google.golang.org/grpc

### TODO

- add dssa service to create symbolic link
- an exist service may be needed in dssa for efficiency (or stat should return an explicit error)
- SFTP check authorized keys, cf https://pkg.go.dev/golang.org/x/crypto/ssh#example-ServerConfig
- parse FIXMEs
- dssa/grpc List operation with only Size/Time options if more efficient
- grpc ope version
- secure plugin connection with
  - https://github.com/filosottile/mkcert
  - https://go.dev/src/crypto/tls/generate_cert.go
  - https://github.com/grpc/grpc-go/tree/master/examples/features/encryption
- defer shutdown server and wait plugin
- ctrl/c signal for server
- mTLS: test with simple auto-generated certs (without CA)
