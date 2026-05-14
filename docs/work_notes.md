# vdasync

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

generate certificates
- gen test CA
- gen server and or client cert/key

server
- internally gen server cert/key

service
- health-check client

testing

### TLS for service and its plugins

service loads config with plugins config
it calls plugin with related config, or, if not set, with provided args

  -ca string
        server or plugin TLS certificate CA
        used by client to check
        => plugin: as -clientca
  -cert string
        server or plugin TLS certificate
        => plugin: ONLY
  -clientca string
        client TLS certificate CA
        used by server or plugin to check
        => plugin: NO
  -clientcert string
        client TLS certificate
        => plugin: NO
  -clientkey string
        client TLS certificate key
        => plugin: NO
  -insec
        don't check certificate when communicating with server
        => plugin: NO
  -insecplugin
        don't check certificate when communicating with plugins
        => plugin: NO, not necessary
  -key string
        server or plugin TLS certificate key
        => plugin: ONLY
  -notls
        insecure communication with servers over http
        => plugin: NO
  -notlsplugin
        insecure communication with plugins over http
        => plugin: as -notls
