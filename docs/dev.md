# Vdasync development

## Building

On linux amd64 platforms, clone the git repository, install Go development tools and run `make build`.

On other platforms, adapt the `build` rule in the Makefile.

Go is also able to build binaries for any target platform, just add required *build rules in the Makefile.

## Design

### Golang API

The [go API](../dssa/dssa.go) sees any data store through the following simple interface:

- List to retrieve directory entries
- Stat to retrieve entry status like size, permissions and modification time
- Get to read the content of a non-directory entry
- Mkdir to create a new directory entry
- SetStat to change the permissions and modification time of an entry
- Put to write the content of a non-directory entry
- Rm to remove an entry
- Symlink to create a symbolic link
- Checksum to compute the file checksum the most efficient way given the plugin

### gRPC API

A [gRPC API](../grpc/dssa.proto) providing the same kind of interface as the Golang one is provided.

Both remote access and plugin access use the same gRPC API.
A plugin may therefore be implemented with any language supported by gRPC.
