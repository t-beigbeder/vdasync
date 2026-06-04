# vdasync

Vdasync stands for versatile data access and synchronization tool.

Vdasync provides access to files or data, either local or remote,
through a CLI and a simple API.

The CLI main use is to synchronize data among different locations,
for instance for data backup and restore or replication.
Synchronization may be very fast as the implementation leverages concurrency.

Remote access is provided through [gRPC](https://grpc.io/) that requires a HTTP/2 transport between hosts.

Beyond local and remote files access, various data access means may be implemented through the use of plugins.
The tool provides the following plugins:

- object storage through an S3 API,
taking care of OS files attributes (type, permissions and modification time) when synchronizing if needed
- remote files access over SFTP through a sftp client
- client-side encrypted storage over any kind of data access mean: files or plugin

There is also a test plugin that barely provides access to local files.

Plugins are implemented as local gRPC servers, using the same API as for remote access.
A plugin may therefore be implemented with any language supported by gRPC.
It could even be run remotely if it made sense.

## Features

- CLI for synchronization
- gRPC server for remote access
- go API and gRPC API
- plugins
  - S3 storage plugin with simple OS files attributes management
  - SFTP client
  - client-side encryption
- Utility to generate testing TLS certificates for CAs, clients and servers

## Status

Beta status

- feature complete, except `vdasync` exclusion and inclusion lists
- rather good test coverage, mainly missing tests for I/O errors

## Usage

Utilities arguments meaning can be retrieved with `<cli-command> -help`.

CLI tools access to data through DSS, DSS stands for data storage system:
this refers either simply to local or remote files,
or else to a specific configuration of a plugin provided through a configuration file as explained below.

Security of the communications may be lowered or disabled using self-generated certificates or HTTP without TLS.
Such settings are disabled by default and should only be used for testing purposes and with good understanding of the risks induced.

### `vdasync` utility

`vdasync` concurrency is disabled by default, but increasing it is generally recommended to gain better performance,
see details below.

Basic usage is

    vdasync [-dryrun] [-rm] [-check] -source <source DSS> -target <target DSS>

Source and target directories must exist in the case of files, their respective sub-trees will be synchronized.

For instance

    vdasync -dryrun -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -rm -source /path/to/dev -target /path/to/backup/for/dev
    vdasync -dryrun -check -source /path/to/dev -target /path/to/backup/for/dev

Remote access to a `vdaserver` (see below) would be enabled with the following DSS syntax:

    dss://<server>:<port>/path/to/remote

For instance restoring local files from a remote backup:

    vdasync -rm -source dss://backup-server:9443/path/to/backup -target /path/to/dev

Using a plugin is enabled through a configuration file, for instance:

    # file /path/to/s3Config.yaml
    pluginsOptions:
      noTls: true  # for testing only!
    pluginReadyRetries: 4
    pluginReadyTimeout: "100ms"
    plugins:
    - name: s3
      type: vdas3
      addArgs:
      - "-s3profile"
      - test-profile
      - "-s3bucket"
      - test-bucket
      - "-s3prefix"
      - vdasync/tests/backup/dev

The data served by the plugin is accessed through its name, for instance making a backup to S3 object storage:

    vdasync -conc 16 -rm -source /path/to/dev -target s3+dss:/ -config /path/to/s3Config.yaml

This will automatically run the `vdas3` executable plugin (its type above) with the "-s3*" arguments provided in the configuration file above,
here enabling up to 16 concurrent I/O requests.

It should be noted that omitting the `//<server>:<port>` part in the DSS URL means accessing `localhost`
on a dynamically allocated TCP port, which is generally what is wanted for a plugin.
Concerning the `path`part in the URL, it is set to "/" in the case of S3,
as the prefix to use in the bucket is rather provided with the "-s3prefix" argument.
Further details about the `vdas3` plugin are provided below.

### Use of concurrency

As said above, increasing `vdasync` concurrency is generally recommended to gain better performance.
Its setting depends on the infrastructure and the plugins involved.

- As a default, the number of available CPU cores can be provided in many cases.
- Writing to slow devices should reduce it or even disable it (USB stick),
as parallel writes may even become counterproductive.
- Using S3 and other HTTP-based services could require increasing it because related requests involve network latency
but may be run safely in parallel; nevertheless this impacts network resources
and should be balanced with such concern.
- Same remark applies in the case of network based storage like NFS or NAS.
- Access to remote resources must also take care of the target service capacity
that is often shared between many users.
- Client based encryption requires local compute resources,
therefore concurrency will be tuned according to related capacity.

### Plugins configurations and DSS naming


### TLS configuration

### Remote server

### S3 storage simple plugin

### SFTP plugin

### Client-side encryption simple plugin

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
- Symlink to create a symbolic link

### gRPC API

A gRPC API providing the same kind of interface as the Golang one is provided.

Both remote access and plugin access use the same gRPC API.
A plugin may therefore be implemented with any language supported by gRPC.

### Concurrency

The synchronization tooling leverages Golang concurrency features,
enabling fast data access through parallelization of I/O and data processing,
as soon as the infrastructure allows it.

Concurrency may be tuned or even switched off to keep resource usage
as efficient as wanted.
