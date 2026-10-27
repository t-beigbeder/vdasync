# Vdasync operation logs

Operation logs record events on source and target files
during their processing for synchronization

- to enable efficient operations at scale
- to provide reports for traceability

The `vdasync` CLI leverages operation logs on option to enable those features.
This page provides information specific to this mode of operation.

Documentation is a work in progress.

Operation logs features are in beta release.

## General operation

Using operation logs to support a synchronization run is an option
of the `vdasync` command activated by the use of the `-oplcmd` flag.
When operation logs are in use, each I/O operation on a source or target DSS is tracked in the logs,
where they are cached and may avoid repeating the same when the synchronization is applied several times.

Status and events from the operation logs can be exported in CSV format
and can thus support various traceability requirements.

The tool's behaviour may appear counterintuitive for the two following topics:

### Operation log and dryrun

The dryrun mode when activated with the related flag has the effect of avoiding any modification
to the target data storage.

The operation log on the other hand has to be updated to cache and track related read operations
on both the source and the target data storage.

### Operation logs and log files

The standard log files, as configured with the flags

    -level string
        log level, defaults to INFO
    -log string
        log file, defaults to vdasync-<pid>.log in temp dir, "std[out|err]" are known keywords

are not to be confused with operation logs.
Operation logs are stored in files as well but are specific to long running operation and related traceability.
The operation logs files are configured with the flags

    -oplpath string
            operation logs master file
    -oplsto string
            operation logs storage: aio|pbf (default "aio")
    -oplsync string
            operation logs synchronization to disk, periodicity and number of backups seconds:backups (default "600:5")

## Usage
