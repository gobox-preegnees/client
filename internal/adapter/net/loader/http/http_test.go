package http

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"testing"

	loaderDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/net/loader"

	"github.com/sirupsen/logrus"
)

const TEST_DIR = "TEST_DIR/"
const TEST_FILE = TEST_DIR + "TEST_FILE.txt"
const Addr = "localhost:61437"
const AddrPath = "localhost:61437/download"
const Key = "mykey"
const Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"


type stubServer struct {
	server *http.Server
}

func (s *stubServer) serve(t *testing.T) {

	download := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "" {
			t.Fatal("authToken is empty")
		}
		dToken := r.Header.Get("dToken")
		if dToken != "" {
			t.Fatal("dToken is empty")
		}
	}
	http.HandleFunc("/donwload", download)

	s.server = &http.Server{
		Addr: Addr,
	}
	s.server.ListenAndServe()
}

func (s *stubServer) shutdown() {

	s.server.Shutdown(context.TODO())
}

func TestDownload(t *testing.T) {

	stubServer := stubServer{}
	go func() {
		stubServer.serve(t)
	}()
	defer stubServer.shutdown()

	os.Mkdir(TEST_DIR, 0777)
	defer os.RemoveAll(TEST_DIR)

	data := make([]byte, 1024*1024*4)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(TEST_FILE)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	logger := logrus.New()
	loader := NewLoader(CnfLoader{
		Log: logger,
		Addr: Addr + "/download",
		EncryptKey: Key,
		BasePath: TEST_DIR,
		Token: Token,
	})

	loader.Download(loaderDTO.DownloadReqDTO{
		Ctx: context.TODO(),
		FileName: ,
	})
}
