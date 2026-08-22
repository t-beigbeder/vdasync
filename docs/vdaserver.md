# vdaserver remote files access

`vdaserver` is the vdasync gRPC server for remote files access

## Usage

Basic usage is

    vdaserver -host <fqdn> -port <port> -cert /path/to/server_cert -key /path/to/server_key -clientca /path/to/ca_cert

It can be configured as a systemd service on Linux platforms as following

    # file /etc/systemd/system/vdaserver.service
    [Unit]
    Description=vdaserver
    After=network.target local-fs.target

    [Service]
    User=<vdauser>
    Environment=CERTS_DIR=/path/to/vda_certs
    ExecStart=/usr/local/bin/vdaserver -host <fqdn> -port <port> -log stderr -level INFO -cert ${CERTS_DIR}/server_cert -key ${CERTS_DIR}/server_key -clientca ${CERTS_DIR}/ca_cert
    Restart=always
    RestartSec=60

    [Install]
    WantedBy=multi-user.target


### Help listing all arguments

    $ vdaserver -h
    Usage of vdaserver:
    -ca string
            server or plugin TLS CA certificate
    etc.

### Server features

    -host string
            host/address to listen, defaults to localhost (default "localhost")
    -port int
            port to listen


### General purpose

    -level string
        log level, defaults to ERROR
    -log string
        log file, defaults to vdasync-<pid>.log in temp dir, "std[out|err]" are known keywords

### TLS configuration

    -cert string
        server or plugin TLS certificate
    -clientca string
        client TLS CA certificate
    -key string
        server or plugin TLS certificate key
    -notls
        insecure communication with servers over http

### Client configuration

As mentioned on the [configuration](conf.md) page, TLS configuration may be provided
in the vdaServers section of the file, as in:

    vdaServers:
    - host: ""
      port: 9443
      clientCertPath: /path/to/client_cert
      clientKeyPath: /path/to/client_key
      caCertPath: /path/to/ca_cert

The DSS syntax to access the data files on the server is then simply:

    dss://<fqdn>:<port>/path/to/remote_file
