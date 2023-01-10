package encryption

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

var ErrCiphertextTooShort = errors.New("ciphertext too short")

// encryptor.
type encryptor struct {
	key        []byte
	aead       cipher.AEAD
	encryption bool
}

// CnfEncrypter. key required
type CnfEncrypter struct {
	Key       string
	Ecryption bool
}

// NewEncryptor.
// Create a new encryption cipher. Your key will be hashed (256bit). Algorithm: chacha20poly1305
func NewEncryptor(cnf CnfEncrypter) (*encryptor, error) {

	if !cnf.Ecryption {
		return &encryptor{}, nil
	}

	h := sha256.New()
	h.Write([]byte(cnf.Key))
	hashSum := h.Sum(nil)

	key := make([]byte, len(hashSum))
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	return &encryptor{
		key:        key,
		aead:       aead,
		encryption: cnf.Ecryption,
	}, nil
}

// Encrypt.
// Accept data byte -> return decrypted data, err: rand.Read
func (e encryptor) Encrypt(data []byte) ([]byte, error) {

	if !e.encryption {
		return data, nil
	}

	nonce := make([]byte, e.aead.NonceSize(), e.aead.NonceSize()+len(data)+e.aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	encryptedMsg := e.aead.Seal(nonce, nonce, data, nil)
	return encryptedMsg, nil
}

// Decrypt.
// Accept data byte -> return decrypted data, err: 1) ErrCiphertextTooShort, 2) aead.Open
func (e encryptor) Decrypt(data []byte) ([]byte, error) {

	if !e.encryption {
		return data, nil
	}

	if len(data) < e.aead.NonceSize() {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:e.aead.NonceSize()], data[e.aead.NonceSize():]

	out, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}
