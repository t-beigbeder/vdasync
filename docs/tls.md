# Vdasync TLS configuration

## TLS use among vdasync components

gRPC communications with remote servers and with the plugins on localhost
need to be encrypted and authenticated to ensure security.
gRPC authentication basically provides standard TLS authentication using client-side certificates:
[mTLS](https://en.wikipedia.org/wiki/Mutual_authentication#mTLS).

This is the model applied for securing communications between vdasync's CLI clients and remote `vdaserver`,
or the different plugins on localhost.

While not recommended, using self-signed certificates can be requested, it disables client authentication.
Disabling TLS completely may also be explicitely requested.

## Testing certificates generator arguments

A testing certificates generator is provided: `testcerts`.
The usage of its arguments following is explained after.

    $ testcerts -h
    Usage of bin/lamd64/testcerts:
    -ca string
            server or plugin TLS CA certificate
    -cakey string
            TLS CA certificate key
    -cert string
            server or plugin TLS certificate
    -clientca string
            client TLS CA certificate
    -clientcert string
            client TLS certificate
    -clientkey string
            client TLS certificate key
    -cn string
            Common name of the TLS CA (default "CA")
    -host string
            TLS certificate host for self-signed (default "localhost")
    -hosts string
            List of TLS certificate hosts, separated by comma, if empty: client certificate
    -key string
            server or plugin TLS certificate key

## Testing certificates generator usage

`testcerts` generates private keys and certificates for:

- self-signed certificates
- CA, always self-signed
- servers certificates for their FQDNs signed by a given CA
- client certificates signed by a given CA

While testing certificates are not recommended for production use,
the following samples leverage `testcerts` as a mean to provide explicit and simple explanations.

A TLS client always authenticates the server requested FQDN for an approved list of CAs, in that case the server CA.
mTLS server will in addition authenticate the client for an approved client CA.
Because the CAs are self-signed and not official ones, their certificates must be provided to the `vdasync` components:

- `-clientca` on the server side to authenticate their clients
- `-ca` on the client side to authenticate the server

An alternative to CLI arguments is to use configuration files, must less verbose, see related section.

Plugins use the same CA than the client which activates them, provided in the CLI configuration file.

CA files generation is achieved for instance with

	testcerts -ca sca-cert.pem -cakey sca-key.pem -cn Server-CA

Corresponding server files, both valid for `localhost` and `some-fqdn` are generated with:

    testcerts -ca sca-cert.pem -cakey sca-key.pem \
      -hosts localhost,some-fqdn -cert some-fqdn-cert.pem -key some-fqdn-key.pem

Doing the same for a client, omitting the hosts argument generates a client certificate:

    testcerts -ca cca-cert.pem -cakey cca-key.pem -cn Client-CA
    testcerts -ca cca-cert.pem -cakey cca-key.pem \
      -cert some-client-cert.pem -key some-client-key.pem

Plugins running on localhost will also use a certificate generated from the same client CA:

	testcerts -ca cca-cert.pem -cakey cca-key.pem -hosts localhost \
    -cert plugin-cert.pem -key plugin-key.pem

## Servers, clients and plugins configuration

Copying those files in the working directories of clients, plugins and servers, this will give for instance:

    vdaserver -host some-fqdn -port 9443 \
      -cert some-fqdn-cert.pem -key some-fqdn-key.pem \
      -clientca cca-cert.pem
    vdasync [...] -target dss://some-fqdn:9443/dir \
      -clientcert some-client-cert.pem -clientkey some-client-key.pem \
      -ca sca-cert.pem

Concerning the plugins configuration, using a configuration file such as the following
for the testing plugin `localFiles` (see [localFiles plugin](localfiles.md))

    # file tlsConfig.yaml
    pluginsOptions:
      certPath: /path/to/plugin-cert.pem
      keyPath: /path/to/plugin-key.pem
      caCertPath: /path/to/cca-cert.pem
    plugins:
    - name: lfs
      type: localFiles

We can check the TLS configuration as following,
this time the server CA is used to authenticate the plugin
and it is provided in the configuration file

    vdasync [...] -target lfs+dss:/dir \
      -config tlsConfig.yaml \
      -clientcert client-cert.pem -clientkey client-key.pem
