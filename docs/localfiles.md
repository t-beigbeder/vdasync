# localFiles plugin

## The `localFiles` test plugin

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
