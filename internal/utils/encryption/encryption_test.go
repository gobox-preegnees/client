package encryption

import (
	"testing"
)

func TestEncryption(t *testing.T) {

	key := "newKey"
	data := []byte("hello world")
	enc, err := NewEncryptor(CnfEncrypter{
		Key: key,
	})
	if err != nil {
		t.Fatal(err)
	}

	eout, err := enc.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}

	dout, err := enc.Decrypt(eout)
	if err != nil {
		t.Fatal(err)
	}

	if string(dout) != string(data) {
		t.Fatal("string(dout) != string(data)")
	}
}