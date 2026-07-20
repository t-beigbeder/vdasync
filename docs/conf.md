# Vdasync and vdaservice configuration

## Configuration files

The `vdasync` and `vdaservice` arguments are numerous and verbose but may generally be provided in a configuration file:

- arguments are taken in priority to their respective entry in the configuration file
- default configuration file is
[`$XDG_CONFIG_HOME`](https://wiki.archlinux.org/title/XDG_Base_Directory)`/vdasync/config.yml`
- it can be overriden with the environment `$VDASYNC_CONFIG` providing another path
- it can be overriden with `-config /path/to/configFile.yml`

## Yaml configuration

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

See also [TLS page](tls.md)
