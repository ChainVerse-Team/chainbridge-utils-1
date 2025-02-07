// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

/*
The keystore package is used to load keys from keystore files, both for live use and for testing.

# The Keystore

The keystore file is used as a file representation of a key. It contains 4 parts:
- The key type (secp256k1, sr25519)
- The PublicKey
- The Address
- The ciphertext

This keystore also requires a password to decrypt into a usable key.
The keystore library can be used to both encrypt keys into keystores, and decrypt keystore into keys.
For more information on how to encrypt and decrypt from the command line, reference the README: https://github.com/ChainSafe/ChainBridge

# The Keyring

The keyring provides predefined secp256k1 and srr25519 keys to use in testing.
These keys are automatically provided during runtime and stored in memory rather than being stored on disk.
There are 5 keys currenty supported: Alice, Bob, Charlie, Dave, and Eve.
*/
package keystore

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ChainSafe/chainbridge-utils/crypto"
	"github.com/ChainSafe/chainbridge-utils/hash"
	"github.com/awnumar/memguard"
	secp256k1 "github.com/ethereum/go-ethereum/crypto"
)

const EnvPassword = "KEYSTORE_PASSWORD"

var keyMapping = map[string]string{
	"ethereum":  "secp256k1",
	"substrate": "sr25519",
}

// KeypairFromAddress attempts to load the encrypted key file for the provided address,
// prompting the user for the password.
func KeypairFromAddress(addr, chainType, path string, insecure bool) (crypto.Keypair, *memguard.Enclave, error) {
	if insecure {
		return insecureKeypairFromAddress(path, chainType)
	}
	path = fmt.Sprintf("%s/%s.key", path, addr)
	// Make sure key exists before prompting password
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("key file not found: %s", path)
	}

	var pswd []byte
	if pswd == nil {
		pswd = GetPassword(fmt.Sprintf("Enter password for key %s:", path))
	}
	hshPwd, salt, err := hash.HashPasswordIteratively(pswd)
	for i := 0; i < len(pswd); i++ {
		pswd[i] = 0
	}

	if err != nil {
		for i := 0; i  < len(hshPwd); i++ {
			hshPwd[i] = 0
		}
		for i := 0; i < len(salt); i++ {
			salt[i] = 0
		}
		return nil, nil, err
	}

	hshPwd = append(hshPwd, salt...)
	for i := 0; i < len(salt); i++ {
		salt[i] = 0
	}

	kp, key, err := ReadFromFileAndDecrypt(path, hshPwd, keyMapping[chainType])
	for i := 0; i  < len(hshPwd); i++ {
		hshPwd[i] = 0
	}
	if err != nil {
		// destroy the keypair
		kp.DeleteKeyPair()
		kp = nil
		
		return nil, nil, err
	}

	return kp, key, nil
}

// BytesToPrivateKey converts a []byte to *ecdsa.PrivateKey
func BytesToPrivateKey(keyBytes []byte) (*ecdsa.PrivateKey, error) {
	return secp256k1.ToECDSA(keyBytes)
}
