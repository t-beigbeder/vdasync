# vdaservice command

`vdaservice` is an auxiliary CLI used for operations, administration and testing purposes.

## Usage

Notes:

- plugins must be configured through a configuration file, see [configuration page](conf.md).
- TLS configuration is usually provided through a configuration file, see [TLS page](tls.md).
- concurrency is disabled by default, but increasing it is generally recommended, see [vdasync documentation](vdasync.md).

Basic usage is

    vdaservice -dss <target DSS> -cmd <command> [-check] [-recur] [-sort] [-tsort]

DSS stands for data storage system:
this can refer either to local files, to remote files accessed on a host running `vdaserver`,
or else to a plugin configured through a file,
see [DSS syntax](dssurl.md).

For instance

    vdaservice -cmd list -dss /path/to/directory -recur -check

### Help listing all arguments

    $ vdaservice -h
    Usage of vdaservice:
    -ca string
            server or plugin TLS CA certificate
    etc.

### Listing features

    -check
        compute checksums
    -csal string
        comma separated list of hash algoritms to compute checksums: sha256 sha512 sha3_256 sha3_512 (default "sha256")
    -excl string
        file containing regexps for paths to be excluded, defaults to none
    -incl string
        file containing regexps for paths to be included, defaults to all
    -dss string
        target of the command
    -sort
            sort output with entries paths
    -tsort
            sort output with entries modification times

### General purpose

    -conc int
        number of concurrent activities
    -config string
        configuration file, see documentation
    -level string
        log level, defaults to ERROR
    -log string
        log file, defaults to vdasync-<pid>.log in temp dir, "std[out|err]" are known keywords
    -out string
        file for output, defaults to stdout, "std[out|err]" are known keywords (default "stdout")
    -version

### TLS configuration

    -cert string
        server or plugin TLS certificate
    -clientca string
        client TLS CA certificate
    -clientcert string
        client TLS certificate
    -clientkey string
        client TLS certificate key
    -insec
        don't check certificate when communicating with server
    -insecplugin
        don't check certificate when communicating with plugins
    -key string
        server or plugin TLS certificate key
    -notls
        insecure communication with servers over http
    -notlsplugin
        insecure communication with plugins over http
