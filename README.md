# vdasync

Versatile data access and synchronization tool.

Vdasync provides access to files or data, either local or remote,
through a CLI and a simple API.

The CLI main use is to synchronize data among different locations,
for instance for data backup and restore or replication.
Synchronization may be very fast as the implementation leverages concurrency.

Remote access is provided through gRPC that requires a HTTP/2 transport between hosts.

Various data access means may be implemented through the use of plugins.
The tool provides the following plugins:

- object storage through an S3 API,
taking care of OS files attributes (type, permissions and modification time) if needed
- sftp client to access remote files through SFTP
- client-side encrypted storage over any kind of data access mean: files or plugin
- access to local files, a plugin that simulates remote access locally for testing purpose

## Status

Beta status

- go API and gRPC API complete with rather good test coverage,
mainly missing tests for I/O errors
- CLI for synchronization
- gRPC server
- S3 storage plugin with simple OS files attributes management
- SFTP client
- local files testing plugin

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
