# vdaencrypt plugin

## Technical description and use cases

This plugin enables to encrypt files and their metadata using the [age](https://age-encryption.org/) encryption tool and library.
Multiple public keys (`age` recipients) may be provided for the encryption,
so that different users can decrypt the files with their respective private keys (`age` identities).
The resulting data may be stored locally or on a remote [DSS server](vdaserver.md).

Encrypted data may be safely stored on _insecure_ storage (public cloud, unencrypted laptop disk, removable drive)
because the encrypted data identities are kept on client side and only leveraged there.
When used with `vdaserver` remote encrypted files, the server may also be hosted on _insecure_ environments
because it only sees opaque encrypted data while decryption is fully done on the client side.

## Limitations and restrictions

### Scalability and reliability

It is a "simple" encryption tool because local files attributes and directories content are globally

- loaded in client memory during data access
- stored in a single encrypted file updated at the end of the synchronization, keeping previous versions as backup

Such metadata handling has limitations both in terms of scalability (could handle 100k files but not 10 millions),
and reliability (metadata global file update can fail after numerous encrypted data files updates).
The second point is generally not a big concern as synchronization can be run as many times as wanted until it succeeds.
However, such errors can leave unreferenced data files that require periodic clean-up.
Corrupted metadata files may require manual clean-up.

### Restrictions on the remote data storage

Even if the encrypted data is technically available to several users in parallel, such use cases are not managed at all.

Computing checksums on the encrypted data requires downloading the content locally to decrypt it, consuming corresponding bandwidth.

If unacceptable, see the [DSS trusted server](vdatserver.md) component, which computes encryption and manage metadata on the server-side
but requires to upload public and private keys.

To be completed.

## Usage

Plugin is `vdaencrypt` and its arguments can be found with `vdaencrypt -help`. Main arguments are:

- `-ageidf` a file providing the list of `age` identities for encrypting data
- `-agerecf` a file providing the list of `age` recipients for decrypting data
- `-underlying` the DSS URL providing the local or remote files root directory for encrypted files storage

TLS arguments for the communication with the plugin apply as usual.
When using `vdaserver` for remote encrypted files storage, the plugin also acts as a DSS client and related TLS options apply:
`-ca` for the server CA, `-clientkey` and `-clientcert` for the client identity, provided here as flags
but more likely as corresponding entries in the plugin configuration file.

### Configuration sample for local encryption

The following configuration

    pluginsOptions:
      clientCertPath: /path/to/client_cert
      clientKeyPath: /path/to/client_key
      certPath: /path/to/plugin_cert
      keyPath: /path/to/plugin_key
      caCertPath: /path/to/ca_cert
    plugins:
    - name: enc_sample
      type: vdaencrypt
      addArgs:
      - "-ageidf"
      - /path/to/age.ids
      - "-agerecf"
      - /path/to/age.recs
      - "-underlying"
      - /path/to/encrypted_files_sample

would be leveraged by vdasync/vdaservice with DSS URL:

    enc_sample+dss:/path/to/data_file

### Configuration sample for remote encryption

    pluginsOptions:
      clientCertPath: /path/to/client_cert
      clientKeyPath: /path/to/client_key
      certPath: /path/to/plugin_cert
      keyPath: /path/to/plugin_key
      caCertPath: /path/to/ca_cert
    plugins:
    - name: remote_enc_sample
      type: vdaencrypt
      addArgs:
      - "-ageidf"
      - /path/to/age.ids
      - "-agerecf"
      - /path/to/age.recs
      - "-underlying"
      - dss://vda-server-fqdn:9443/path/to/encrypted_files_sample
      - "-ca"
      - /path/to/ca_cert
      - "-clientcert"
      - /path/to/client_cert
      - "-clientkey"
      - /path/to/client_key
    vdaServers:
    - host: vda-server-fqdn
      port: 9443
      clientCertPath: /path/to/client_cert
      clientKeyPath: /path/to/client_key
      caCertPath: /path/to/ca_cert

would be leveraged by vdasync/vdaservice with DSS URL:

    remote_enc_sample+dss:/path/to/data_file
