package common

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

// AgeNewKeyPair generates an age public/private key-pair.
// Returns the public and private key pair ascii encoded, or an error if any occurs.
func AgeNewKeyPair() (string, string, error) {
	xi, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}
	return xi.Recipient().String(), xi.String(), nil
}

// AgeEncryptMsg encrypts a bytes message for the corresponding recipients public keys,
// returns the encoded byres or an error if any occurs.
func AgeEncryptMsg(msg []byte, srs ...string) ([]byte, error) {
	bsa := bytes.Buffer{}
	var rs []age.Recipient
	for _, sr := range srs {
		r, err := age.ParseX25519Recipient(sr)
		if err != nil {
			return nil, fmt.Errorf("in EncryptMsg: %w", err)
		}
		rs = append(rs, r)
	}
	wc, err := age.Encrypt(&bsa, rs...)
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	_, err = io.Copy(wc, bytes.NewReader(msg))
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	err = wc.Close()
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	return bsa.Bytes(), nil
}

// AgeDecryptMsg decrypts an age encrypted message using
// any of the provided ascii encoded private keys
// and returns the resulting bytes message or any error
// if one occurs.
func AgeDecryptMsg(bs []byte, sids ...string) ([]byte, error) {
	var ids []age.Identity
	for _, sid := range sids {
		id, err := age.ParseX25519Identity(sid)
		if err != nil {
			return nil, fmt.Errorf("in DecryptMsg: %w", err)
		}
		ids = append(ids, id)
	}
	rd, err := age.Decrypt(bytes.NewReader(bs), ids...)
	if err != nil {
		return nil, fmt.Errorf("in DecryptMsg: %w", err)
	}
	bss, err := io.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("in DecryptMsg: %w", err)
	}
	return bss, nil
}

// AgeEncrypt encrypts a writer's content to one or more recipients ascii encoded public keys.
//
// Writes to the returned WriteCloser are encrypted and written to dst as an age file.
// Every recipient will be able to decrypt the file.
//
// The caller must call Close on the WriteCloser when done for the last chunk to be encrypted
// and flushed to dst.
func AgeEncrypt(dst io.Writer, srs ...string) (io.WriteCloser, error) {
	var rs []age.Recipient
	for _, sr := range srs {
		r, err := age.ParseX25519Recipient(sr)
		if err != nil {
			return nil, fmt.Errorf("in Encrypt: %w", err)
		}
		rs = append(rs, r)
	}
	return age.Encrypt(dst, rs...)
}

// AgeDecrypt decrypts a reader age-encrypted to one or more identities.
//
// It returns a Reader reading the decrypted plaintext of the age file read from src.
// All identities will be tried until one successfully decrypts the file.
func AgeDecrypt(src io.Reader, sids ...string) (io.Reader, error) {
	var ids []age.Identity
	for _, sid := range sids {
		id, err := age.ParseX25519Identity(sid)
		if err != nil {
			return nil, fmt.Errorf("in Decrypt: %w", err)
		}
		ids = append(ids, id)
	}
	return age.Decrypt(src, ids...)
}
