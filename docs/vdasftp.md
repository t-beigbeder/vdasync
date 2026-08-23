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
