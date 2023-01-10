package http

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	loaderDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/net/loader"
	ecnryption "github.com/gobox-preegnees/gobox-client/internal/utils/encryption"

	"github.com/sirupsen/logrus"
)

// loader.
type loader struct {
	log        *logrus.Logger
	addr       string
	encryptKey string
	basePath   string
	token      string

	dMutex      sync.Mutex
	downloading map[string]context.CancelFunc

	uMutext   sync.Mutex
	uploading map[string]context.CancelFunc
}

// CnfLoader.
type CnfLoader struct {
	Log        *logrus.Logger
	Addr       string
	EncryptKey string
	BasePath   string
	Token      string
}

// NewLoader.
func NewLoader(cnf CnfLoader) *loader {

	return &loader{
		log:        cnf.Log,
		addr:       cnf.Addr,
		encryptKey: cnf.EncryptKey,
		basePath:   cnf.BasePath,
		token:      cnf.Token,

		dMutex:      sync.Mutex{},
		downloading: make(map[string]context.CancelFunc),

		uMutext:   sync.Mutex{},
		uploading: make(map[string]context.CancelFunc),
	}
}

func (l *loader) Download(in loaderDTO.DownloadReqDTO) error {

	// TODO: сделать что нибудь с ошибками
	// TODO: нужно сверху проверить есть ли файл в заблокированных

	l.abortDownloading(in.FileName)

	ctx, cancel := context.WithCancel(in.Ctx)
	l.createDonwloading(cancel, in.FileName)

	l.prepareFile(in.F, in.SizeFile)

	counter := &WriteCounter{
		Log:   l.log,
		Total: 0,
	}
	
	encryptor, err := ecnryption.NewEncryptor(ecnryption.CnfEncrypter{
		Key: l.encryptKey,
	})
	if err != nil {
		return err
	}

	fWriter := &FileWriter{
		Ctx: ctx,
		Log: l.log,
		F:   in.F,
	}
	if l.token != "" {
		fWriter.WithEncryption(encryptor)
	} else {
		fWriter.WithoutEncryption()
	}
	fWriter.stopOnCancel()

	resp, err := l.createRequest(in.DToken)
	if err != nil {
		return err
	}

	_, err = io.Copy(fWriter, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// TODO: сверху нужно закрыть файл
	return nil
}

// abortDownloading. 
// Accept fileName (path) -> cancelFun() current downloading if exists
func (l *loader) abortDownloading(fileName string) {

	l.dMutex.Lock()
	cancel, ok := l.downloading[fileName]
	if ok {
		cancel()
	}
	l.dMutex.Unlock()
}

// createDonwloading.
// Accept fileName of file, which need to be downloaded and cacncelFunc for cancel downloading
func (l *loader) createDonwloading(cancel context.CancelFunc, fileName string) {

	l.dMutex.Lock()
	l.downloading[fileName] = cancel
	l.dMutex.Unlock()
}

// prepareFile. 
// Truncate and seek in current file, which need to download
func (l *loader) prepareFile(f *os.File, sizeFile int64) error {

	if err := f.Truncate(sizeFile); err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

// createRequest.
// Accept downloadToken (dToken), create request
func (l *loader) createRequest(dToken string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodGet, l.addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header["Authorization"] = strings.Fields("Bearer " + l.token)
	req.Header["dToken"] = strings.Fields(dToken)

	cli := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *loader) Upload(req loaderDTO.UploadDTO) {

}

func (l *loader) replaceBackSlash() {}
