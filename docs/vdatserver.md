# vdatserver trusted server

`vdatserver` provides remote access to encrypted files:

- supports concurrent updates by several clients,
- periodically saves encrypted metadata on underlying storage and flags inconsistencies
- computes checksums on the server-side from the underlying storage

## General operation

`vdatserver` can be used as an alternative to the `vdaencrypt` client-side encryption plugin
to work around some of its related limitations.

As providing remote encryption/decryption services,
it requires a `vdaservice` client to push encryption public keys and at least one decryption private key.
This can be a security concern when hosted in environments not fully trusted, among others.
This requires obviously to protect gRPC access with proper TLS encryption and authentication.

## Usage

Basic usage is

    vdatserver -host <fqdn> -port <port> -cert /path/to/server_cert -key /path/to/server_key -clientca /path/to/ca_cert -ageidf <identity-file>

Usage is the same as for [vdaserver](vdaserver.md) with the addition of the server decryption key:

    -ageidf string
          age identities (secrets) file name

This private key enables the server to decrypt secret information provided
by `vdaservice` to unlock access to encrypted data as explained below.

Once unlocked, access from the client is the same as for a regular `vdaserver`:

    dss://<trusted-server-fqdn>:<port>/path/to/remote_file

### Unlock/lock access to encrypted data

The `vdaservice` utility is used to unlock or later lock the access
to encrypted data on the server.

Unlocking the access uses the `-cmd trust` flag with the following arguments:

    -ageeidf string
        DSS encryption age identities (secrets) file name
    -ageerecf string
        DSS encryption age recipients (public keys) file name
    -agetrecf string
        trusted server age recipients (public keys) file name

The `-agetrecf` refers to the public key enabling the trusted server to decrypt information with its private key.

The `-ageeidf` and `ageerecf` are the keys used for encrypting/decrypting DSS data
on the underlying storage.
They are sent encrypted to the server with its public key.
Only the trusted server can thus decrypt them.

Locking the access afterwards uses the `-cmd untrust` flag.

## Limitations and restrictions

### Scalability and consistency

It is a "simple" encryption tool because local files attributes and directories content are globally

- loaded in server memory during data access
- stored in a single encrypted file updated periodically,
keeping previous versions as backup

Such metadata handling has limitations both in terms of scalability (could handle 100k files but not 10 millions),
and consistency (metadata global file update can fail after numerous encrypted data files updates).

Inconsistencies are flagged and require analyzing the encrypted metadata
stored under the underlying directory in the `.vdasync.meta` file on one side,
and the encrypted data stored in files distributed under the same underlying directory.
The `vdservice` utility comes with a repair flag to recover from such incidents.
