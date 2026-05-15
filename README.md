# vdasync

Versatile data access and synchronization tool.

Vdasync provides access to files or data, either local or remote,
through a CLI and a simple API.

The CLI main use is to synchronize data among different locations,
for instance for backup and restore, or controlled data replication.
Synchronization may be very fast as the implementation leverages concurrency.

Remote access is implemented with gRPC that uses HTTP/2 transport.

The API also enables an easy implementation of various data access means
through the use of plugins. The tool provides the following plugins:

- object storage through an S3 API,
synchronizing OS files and directories attributes if wanted
- sftp client
- client-side encrypted storage over any kind of data access mean: files or plugin
- testing plugin simply providing access to local files

## Status

Alpha status

- go API and gRPC API complete with rather good test coverage,
missing tests for I/O errors
- CLI for synchronization
- local files testing plugin
- gRPC server

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

Both remote access and plugin access use the same gRPC API.
A plugin may therefore be implemented with any language supported by gRPC.

### Concurrency

The synchronization tooling leverages Golang concurrency features,
enabling fast data access through parallelization of I/O and data processing,
as soon as the infrastructure allows it.

Concurrency may be tuned or even switched off to keep resources usage
as efficient as wanted.
