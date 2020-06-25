# Keycard Crypt

Keycard Crypt is a CLI tool to encrypt / decrypt files using encryption keys generated with Keycard. It uses keys from the [EIP-1581](https://eips.ethereum.org/EIPS/eip-1581) tree. It can process both single files and entire directories recursively. Each file will be encrypted / decrypted with a unique key using the card derivation feature. The [Secure IO](https://github.com/minio/sio) library is used to handle encryption / decryption of files. The encrypted files have .kef extension and will be stored in the same directory as originals.

## Installation

Download Keycard Crypt from the [release page](https://github.com/choppu/keycard-crypt/releases).

## Usage

### Encrypt

`keycard-crypt encrypt paths...`

### Decrypt

`keycard-crypt decrypt paths...`

By default the original files are removed once the encryption / decryption process is completed. To keep the original files use `keycard-crypt -keep-originals encrypt paths...` for encrypt and `keycard-crypt -keep-originals decrypt paths...` for decrypt.





