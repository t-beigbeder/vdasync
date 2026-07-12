# vdasync

Vdasync is a versatile data access and synchronization tool, providing access to local files,
to remote files and to various data access means through the use of plugins.

## General

Vdasync synchronization tool is a [CLI](https://en.wikipedia.org/wiki/Command-line_interface)
running on Go [supported platforms](https://go.dev/wiki/MinimumRequirements).

As [rsync](https://linux.die.net/man/1/rsync),
`vdasync` is intended to be used for backups and mirroring and as an improved copy command for everyday use.
Beyond local and remote files access, synchronization may leverage other data access means through plugins.
`vdasync` also adds the capability to track long-running synchronization operations on large datasets,
through the use of operations logs.

It leverages Go [concurrency features](https://go.dev/wiki/LearnConcurrency)
to enable the use of available resources as efficiently as wanted.

It comes with the following components:

- [vdasync](docs/vdasync.md), the synchronization CLI,
- [vdaserver](docs/vdaserver.md), a [gRPC](https://grpc.io/) server providing remote access to files,
- [vdaservice](docs/vdaservice.md), a CLI mainly used for administration and testing purposes,
- a gRPC plugin API to add other data access means, among which the following are provided:

  - [vdas3](docs/vdas3.md), storing data as S3 storage objects,
  - [vdasftp](docs/vdasftp.md), accessing files available from on a SFTP server,
  - [vdaencrypt](docs/vdaencrypt.md), storing encrypted data on local or remote files,

## Deployment overview

## Basic usage

## Status

Vdasync is actively tested on Linux amd64 and tested on Windows amd64.