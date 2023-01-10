package http

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type IEncryprot interface {
	Decrypt(data []byte) ([]byte, error)
}

// FileWriter. Write to a file
type FileWriter struct {
	Ctx       context.Context
	Log       *logrus.Logger
	F         *os.File

	encryptor IEncryprot
	withEnctyption bool
	
	stopped        bool
}

// WithEncryption. With encryption mode
func (f *FileWriter) WithEncryption(encryptor IEncryprot) *FileWriter {

	f.encryptor = encryptor
	f.withEnctyption = true
	return f
}

// WithoutEncryption. None encryption mode
func (f *FileWriter) WithoutEncryption() *FileWriter {

	f.withEnctyption = false
	return f
}

// stopOnCancel. Watch context cancel
func (f *FileWriter) stopOnCancel() {

	go func() {
		select {
		case <-f.Ctx.Done():
			f.stopped = true
		}
	}()
}

func (f *FileWriter) Write(p []byte) (int, error) {

	if f.stopped {
		return 0, nil
	}
	
	if f.withEnctyption {
		data, err := f.encryptor.Decrypt(p)
		if err != nil {
			return len(p), err
		}
		return f.F.Write(data)
	}
	return f.F.Write(p)
}
