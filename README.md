# vdasync

Versatile data access and synchronization tool.

Vdasync provides access to files or data, either local or remote,
through a CLI and a Golang simple API.
The CLI main use is to synchronize data among different locations,
for instance for backup and restore.
Synchronization may be very fast as the implementation leverages concurrency.

The API also enables an easy implementation or various data access means
through the use of plugins. The tool provides the following plugins:

- object storage through an S3 API
- sftp client
- client-side encrypted storage on any kind of storage: files or plugin

## Status

Alpha status

- go API and gRPC API complete with rather good test coverage, missing I/O errors
- CLI for synchronization with plugin to simulate gRPC server on localhost, both using local filesystem

## Design

### Golang API

The Golang API sees any data store through the following simple interface:

- List to retrieve directory entries
- Stat to retrieve entry status like size, permissions and modification time
- Get to read the content of a non-directory entry
- Mkdir to create a new directory entry
- SetStat to change the permissions and modification time of an entry
- Put to write the content of a non-directory entry
- Rm to remove an entry
- Symlink to create a symbolic linl

### gRPC API

A gRPC API providing the same kind of interface as the Golang one is provided.

Both remote access and plugin access use this API.
Thus a plugin may be implemented with any language supported by gRPC.

### Concurrency

The synchronization tooling leverages Golang concurrency features,
thus enabling very fast data access through parallelization of I/O
and data processing, if the infrastructure permits.
