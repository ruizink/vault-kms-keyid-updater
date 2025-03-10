# HashiCorp Vault KMS KeyID Updater

## Description

[![Build Status](https://github.com/ruizink/vault-kms-keyid-updater/actions/workflows/build.yml/badge.svg)](https://github.com/ruizink/vault-kms-keyid-updater/actions/workflows/build.yml)

A simple tool for HashiCorp Vault to update the KMS KeyID inside `vault.db`.

The use case for this rather specific, but it can be useful in situations where you're using one of the supported KMS systems for auto-unseal, combined with BYOK (Bring your own key).

Imagine a scenario where you need to create a new Managed Vault (Azure Key Vault, AWS KMS or GCP Cloud KMS), and then reimport your own auto-unseal key.

When importing the key, a new version number for that key will be automatically created by the provider. Since HashiCorp Vault keeps this version number (as KeyID) stored in the database, that means that if you lose your existing KMS instance, you won't be able to unseal Vault despite knowing the unseal key, because Vault will be pointing to a key version that does not exist in the newly created KMS instance.

You would see this type of error in the logs:

```text
error fetching stored unseal keys failed: failed to decrypt keys from storage: error decrypting seal wrapped value\nerror decrypting using seal azurekeyvault: POST https://<redacted_akv>.vault.azure.net/keys/<redacted_key_name>/<redacted_key_version>/unwrapkey
```

With this tool, you can change the KeyID to the version that was auto-generated in the new KMS when importing your own key.

## Usage example

To update the KeyID from `<old_key>` to `<new_key>` you need to update both `core/hsm/barrier-unseal-keys` and `core/recovery-key` BoltDB keys, from the `data` bucket.

You can do this by running:

```bash
./vault-kms-keyid-updater -db /opt/vault/raft/vault.db -bucket data -boltkey core/hsm/barrier-unseal-keys -keyid "<old_key>" -newkeyid "<new_key>"
./vault-kms-keyid-updater -db /opt/vault/raft/vault.db -bucket data -boltkey core/recovery-key -keyid "<old_key>" -newkeyid "<new_key>"
```
