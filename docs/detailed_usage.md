# vdasync's detailed usage

## Detailed usage

To be completed

### Plugins configurations and DSS naming

### TLS configuration

gRPC communications with remote servers and even with the plugins on localhost need to be encrypted and authenticated to enforce security.
gRPC authentication may be customized in many ways but basically provides standard TLS authentication using client-side certificates:
[mTLS](https://en.wikipedia.org/wiki/Mutual_authentication#mTLS).

The is the model applied by default for securing communications between vdasync's components: CLI clients towards remote vdaserver, or towards different plugins on localhost.

While not recommended, using self-signed certificates (no client authentication) and even disabling TLS may be explicitely activated.

A testing certificates generator is provided: `testcerts`. It generates private keys and certificates for:

- self-signed certificates
- CA, always self-signed
- servers  certificates for their FQDNs from a given CA
- client certificates from a given CA

While testing certificates are not recommended for production use,
the following samples leverage `testcerts` for providing explicit explanations.

TLS client always authenticate server FQDN for an approved list of CAs, in that case the server CA.
mTLS server will in addition authenticate the client for an approved client CA.
Because the CAs are self-signed and not official ones, their certificates must be provided to the vdasync components:

- `-clientca` on the server or plugin side
- `-ca` on the client side

CA files generation is achieved for instance with

	testcerts -ca sca-cert.pem -cakey sca-key.pem -cn Server-CA

Corresponding server files, both valid for localhost and some-fqdn are generated with

    testcerts -ca sca-cert.pem -cakey sca-key.pem \
      -hosts localhost,some-fqdn -cert some-fqdn-cert.pem -key some-fqdn-key.pem

Doing the same for a client (no hosts argument)

    testcerts -ca cca-cert.pem -cakey cca-key.pem -cn Client-CA
    testcerts -ca cca-cert.pem -cakey cca-key.pem \
      -cert some-client-cert.pem -key some-client-key.pem

### Remote server

### S3 storage simple plugin

### SFTP plugin

### Client-side encryption simple plugin

### Logging information
