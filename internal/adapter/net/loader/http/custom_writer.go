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
	Ctx context.Context
	Log *logrus.Logger
	F   *os.File

	Encryptor      IEncryprot
	WithEnctyption bool

	stopped bool
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
		return 0, context.Canceled
	}

	if f.WithEnctyption {
		data, err := f.Encryptor.Decrypt(p)
		if err != nil {
			return len(p), err
		}
		return f.F.Write(data)
	}
	return f.F.Write(p)
}
