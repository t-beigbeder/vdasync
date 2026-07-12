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

The schema below illustrates how `vdasync` synchronizes
local or remote files in a standalone or client-server deployment.

![Vdasync's local and remote deployments](docs/images/vdasync-deployment.png "Vdasync's local and remote deployments schema")

Using cloud or network storage services such as S3 object storage or SFTP
is enabled with plugins as shown below:

![Vdasync's plugins deployments](docs/images/vdasync-plugins.png "Vdasync's plugins deployments schema")

## Basic usage

Basic usage is

    vdasync [-dryrun] [-rm] [-check] -source <source DSS> -target <target DSS>

DSS stands for data storage system:
this can refer either simply to local files, to remote files accessed on a host running `vdaserver`,
or else to a plugin configured through a file.
Use `vdasync -help` to display the flags and their meanings.
See the dedicated [page](docs/vdasync.md) for detailed information.

Source and target directories must exist in the case of files, their respective sub-trees will be synchronized.

For instance

    vdasync -dryrun -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -dryrun -check -source /path/to/dev -target /path/to/backup/for/dev

Remote access to a `vdaserver` would be enabled with the following DSS syntax:

    dss://<server>:<port>/path/to/remote

For instance restoring local files from a remote backup:

    vdasync -rm -source dss://backup-server:9443/path/to/backup -target /path/to/dev

## Status

Vdasync is actively tested on Linux amd64 and tested on Windows amd64.