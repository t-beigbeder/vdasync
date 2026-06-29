# vdasync's detailed usage

## Detailed usage

This page details how the `vdasync` tool and related components work,
and how to control them with command-line arguments and configuration.

### DSS naming

DSS names are URL formatted as following:

- `relativePath`: access local files relative to working directory
- `/absolute/to/path`: access local files with absolute path
- `pluginName+dss:/`: access data through _pluginName_, plugins define their root directory (or equivalent) as arguments,
so DSS names use an empty path (with the exception of the `localFiles` test plugin)
- `dss://host[:port]/remote/path`: access remote files under /remote/path

### The `localFiles` test plugin

Configuring TLS may require some tests.
Concerning the communication from `vdasync` to its plugins, the configuration can be tested using
the `localFiles` test plugin: indeed this TLS configuration is shared by all plugins.
This plugin simply exposes the local filesystem DSS through the gRPC plugin API.

To use it, just set up the default configuration
`$XDG_CONFIG_HOME/vdasync/config.yml`
with such a content

    pluginsOptions:
      # set TLS as wanted here
    plugins:
    - name: lfs
      type: localFiles
      addArgs: ["-level", "INFO", "-log", "stderr"]

and run a test command:

    vdasync -dryrun -source /path/to/source -target lfs+dss:/path/to/target

Be sure to install the `localFiles` executable in the same directory as `vdasync`.

### Configuration files

The `vdasync` arguments are numerous and verbose but may generally be provided in a configuration file:

- arguments are taken in priority to their respective entry in the configuration file
- default configuration file is
[`$XDG_CONFIG_HOME`](https://wiki.archlinux.org/title/XDG_Base_Directory)`/vdasync/config.yml`
- it can be overriden with the environment `$VDASYNC_CONFIG` providing another path
- it can be overriden with `-config /path/to/configFile.yml`

The yaml configuration format is explained based on an example,
you can consult its format in the [source](../config/config.go).

    pluginsOptions:
      clientCertPath: /local/tmp/certs/client-cert.pem
      clientKeyPath: /local/tmp/certs/client-key.pem
      certPath: /local/tmp/certs/plugin-cert.pem
      keyPath: /local/tmp/certs/plugin-key.pem
      caCertPath: /local/tmp/certs/cca-cert.pem
    plugins:
    - name: lfs
      type: localFiles
    - name: s3
      type: vdas3
      addArgs:
      - "-s3profile"
      - otvl-tests
      - "-s3bucket"
      - otvl-tests
      - "-s3prefix"
      - vdasync/tests/default

This configuration starts two plugins along with the `vdasync` tool.
The plugins use the TLS certificate provided with certPath/keyPath
and authenticate the client's certificate using the CA certificate provided with caCertPath.
The vdasync client authenticates the plugins as TLS servers using the same CA certificate, thus a client CA.
Each plugin receives its own set of additional arguments based on its type,
this is explained in the plugins specific sections.

### TLS configuration

gRPC communications with remote servers and with the plugins on localhost
need to be encrypted and authenticated to ensure security.
gRPC authentication basically provides standard TLS authentication using client-side certificates:
[mTLS](https://en.wikipedia.org/wiki/Mutual_authentication#mTLS).

The is the model applied for securing communications between vdasync's CLI clients and remote `vdaserver`,
or the different plugins on localhost.

While not recommended, using self-signed certificates can be requested, it disables client authentication.
Disabling TLS completely may also be explicitely requested.

A testing certificates generator is provided: `testcerts`. It generates private keys and certificates for:

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


Copying those files in the working directories of clients, plugins and servers, this will give for instance:

    vdaserver -host some-fqdn -port 9443 \
      -cert some-fqdn-cert.pem -key some-fqdn-key.pem \
      -clientca cca-cert.pem
    vdasync [...] -target dss://some-fqdn:9443/dir \
      -clientcert some-client-cert.pem -clientkey some-client-key.pem \
      -ca sca-cert.pem

Concerning the plugins configuration, using a configuration file such as the following
for the testing plugin `localFiles`

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

### Remote server

Launching a remote server will generally be achieved using a `systemd` service,
in which case logging to stderr (see related section) can be preferred.
The main tool arguments are the host (or IP) address of the network interface(s) to listen to,
as well as a reserved TCP port, for instance:

    vdaserver -host some-fqdn -port 9443

`-cert`, `-key` and `-clientca` arguments are provided to secure the communications and authenticate the clients
as explained above:

    vdaserver -host some-fqdn -port 9443 \
      -cert /path/to/some-fqdn-cert.pem -key /path/to/some-fqdn-key.pem \
      -clientca /path/to/cca-cert.pem

Client authentication is disabled by omitting the `-clientca` flag.
TLS is disabled by using the `-notls` flag.

### S3 storage simple plugin

to be completed

### SFTP plugin

to be completed

### Client-side encryption simple plugin

The client-side encryption simple plugin leverages the [`age`](https://github.com/filosottile/age) tool and library.
Files content and metadata (files attributes and directories' content) are encrypted using a list of public keys (`age` recipients)
an decryted using a list of private keys (`age` identities).

Encrypted data may be safely stored on _insecure_ storage (public cloud, unencrypted laptop disk, removable drive)
because the encrypted data identities are kept on client side and only leveraged there.
When used with `vdaserver` remote encrypted files, the server may also be hosted on _insecure_ environments
bacause it only sees opaque encrypted data while decryption is fully done on the client side.

It is a "simple" encryption tool because local files attributes and directories content are globally

- loaded in client memory during data access
- stored in a single encrypted file updated at the end of the synchronization, keeping previous versions as backup

Such metadata handling has limitations both in terms of scalability (could handle 100k files but not 10 millions),
and reliability (metadata global file update can fail after numerous encrypted data files updates).
The second point is generally not a big concern as synchronization can be run as many times as wanted until it succeeds.
However, such errors can leave unreferenced data files that require periodic clean-up.
Corrupted metadata files may require manual clean-up.

Plugin is `vdaencrypt` and arguments can be found with `vdaencrypt -help`. Main arguments are:

- `-ageidf` a file providing the list of `age` identities for encrypting data
- `-agerecf` a file providing the list of `age` recipients for decrypting data
- `-underlying` the DSS URL providing the local or remote files root directory for encrypted files storage

TLS arguments for the communication with the plugin apply as usual.
When using `vdaserver` for remote encrypted files storage, the plugin also acts as a DSS client and related TLS options apply:
`-ca` for the server CA, `-clientkey` and `-clientcert` for the client identity, provided here as flags
but more likely as corresponding entries in the plugin configuration file.

### Logging information

to be completed
