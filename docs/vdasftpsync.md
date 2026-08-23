# vdasftpsync command

`vdasftpsync` integrates the vdasftp plugin inside the vdasync CLI.
This removes the burden of configuring the communication with the sftp plugin
and deploying it when it is the only one in use.

Its usage is the same as [vdasync](vdasync.md).

The sftp arguments declared with the [vdasftp](vdasftp.md) configuration are in that case
provided in a `sftpServers` section in the [configuration](conf.md) file.

For instance:

    sftpServers:
    - name: "sftp_sample"
      user: sftp-login
      address: "otvl-sftp-server:22"
      ident: /home/guest/.ssh/id_ssh_test
      root: /path/to/sftp_root
      knownHostsFile: /home/guest/.ssh/special_kown_hosts


would be leveraged by vdasync/vdaservice with DSS URL using an underscore
in front of the name of the sftpServer configuration entry:

    _sftp_sample+dss:/path/to/data_file
