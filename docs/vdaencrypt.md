# vdaencrypt plugin

This plugin enables to encrypt files and their metadata using the [age](https://age-encryption.org/) encryption library.
Multiple public keys may be provided for the encryption, so that different users can decrypt the files with their respective private keys.
The resulting data may be stored locally or on a remote [DSS server](vdaserver.md).
The encryption public and private keys are never exposed to the remote DSS server if applicable.

Restrictions on the remote data storage:

- even if the encrypted data is technically available to several users in parallel, such use cases are not managed at all,
- computing checksums on the encrypted data requires downloading the content locally to decrypt it, consuming corresponding bandwidth,

for such use cases, see the [DSS trusted server](vdatserver.md) component, which computes encryption and manage metadata on the server-side
at the cost of having to upload public and private keys.

To be completed.