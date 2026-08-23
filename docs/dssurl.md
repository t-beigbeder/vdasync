# Vdasync DSS URLs syntax

DSS names are URLs formatted as following:

- `relativePath`: access local files relative to working directory
- `/local/path`: access local files with absolute path
- `pluginName+dss:/plugin/path`: access data through _pluginName_ under plugin path,
plugins define their root directory (or equivalent) as arguments
- `dss://host[:port]/remote/path`: access remote files under /remote/path

When a plugin name is prefixed with an underscore, it refers to a plugin
integrated with `vdasync` for convenience, today only `vdasftpsync` as in:

    _sftp_sample+dss:/path/to/data_file

referring to the configuration entry:

    sftpServers:
    - name: "sftp_sample"
      # etc.
