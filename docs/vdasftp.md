# vdasftp plugin

This plugin is a SFTP client that provides access to files available on a SFTP server.

See also the [`vdasftpsync`](vdasftpsync.md) CLI for simpler integration.

![vdasftp plugin deployment](images/vdasync-sftp.png "vdasftp plugin deployment schema")

## Usage

Plugin is `vdasftp` and its arguments can be found with `vdasftp -help`. Main arguments are:

    -sftpaddress string
            SFTP server address (default "localhost:22")
    -sftpident string
            SSH identity file to authenticate
    -sftpkhfile string
        known_hosts file, defaults to $HOME/.ssh/known_hosts
    -sftpnohkc
            ignore host key, insecure, equivalent of ssh StrictHostKeychecking=no
    -sftproot string
            root path from SFTP server root where files are served
    -sftpuser string
            SFTP server login
    -conc int
            number of concurrent activities

The concurrency argument enables using as many SFTP clients plus one,
that may be used in parallel when uploading or downloading data files
during configuration.

### Configuration sample

The following configuration

    pluginsOptions:
      clientCertPath: /path/to/client_cert
      clientKeyPath: /path/to/client_key
      certPath: /path/to/plugin_cert
      keyPath: /path/to/plugin_key
      caCertPath: /path/to/ca_cert
    plugins:
    - name: sftp_sample
      type: vdasftp
      addArgs:
    - "-sftpuser"
    - sftp-user
    - "-sftpaddress"
    - otvl-sftp-server:22
    - "-sftpident"
    - /home/guest/.ssh/id_ssh_test
    - "-sftproot"
    - /path/to/sftp_root

would be leveraged by vdasync/vdaservice with DSS URL:

    sftp_sample+dss:/path/to/data_file

## Restrictions

Compared to native OS access to files, the SFTP implementation at least on Linux
is not able to set the last modification time of symbolic links.
The synchronization flag `-nomtlink` can be used to ignore
such differences on symbolic links.
