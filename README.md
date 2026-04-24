# otvl_dtacsy

Versatile data access and synchronization tool

## Design

### Dssa: data storage system abstraction

- local files
- grpc server to local files or any plugin
- plugin through grpc
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

### First steps

- List impl. local, grpc server and plugin (last just for POC)
- full local and grpc server implementation

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

- SFTP check authorized keys, cf https://pkg.go.dev/golang.org/x/crypto/ssh#example-ServerConfig
- internal/remote/lfsvr.go local refactoring to generalized Dssa implem
(so accessible in remote) would be to consider
- secure plugin connection with
  - https://go.dev/src/crypto/tls/generate_cert.go
  - https://github.com/denji/golang-tls
  
