package http

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	// loaderDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/net/loader"
	ecnryption "github.com/gobox-preegnees/gobox-client/internal/utils/encryption"

	"github.com/sirupsen/logrus"
)

const TEST_DIR = "TEST_DIR" + string(filepath.Separator)
const TEST_FILE = TEST_DIR + "TEST_FILE.txt"
const Addr = "localhost:61437"
const AddrPath = "http://localhost:61437/download"
const Key = "mykey"
const Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
const DToken = "fake dToken"

var files = make([]*os.File,0)

type stubServer struct {
	server *http.Server
}

func (s *stubServer) serve() {

	download := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
		}
		dToken := r.Header.Get("dToken")
		if dToken == "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
	http.HandleFunc("/download", download)

	f := createFile(&testing.T{}, 4*1024*1024, TEST_DIR + "FakeFIle")
	defer f.Close() // TODO: почему то файл в конце не удаляется

	getFile := func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(w, f); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	http.HandleFunc("/getFile", getFile)

	s.server = &http.Server{
		Addr: Addr,
	}
	s.server.ListenAndServe()
}

func (s *stubServer) shutdown() {

	s.server.Shutdown(context.TODO())
}

func TestMain(m *testing.M) {

	stubServer := stubServer{}
	go func() {
		stubServer.serve()
	}()
	defer stubServer.shutdown()
	os.Mkdir(TEST_DIR, 0777)
	defer os.RemoveAll(TEST_DIR)

	defer func() {
		for _, v:= range files {
			v.Close()
		}
	}()

	out := m.Run()
	if out != 0 {
		panic(out)
	}
}

func TestCreateRequest(t *testing.T) {

	enc, err := ecnryption.NewEncryptor(ecnryption.CnfEncrypter{
		Key:       Key,
		Ecryption: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	loader := NewLoader(CnfLoader{
		Log:       logger,
		Addr:      AddrPath,
		BasePath:  TEST_DIR,
		Token:     Token,
		Encryptor: enc,
	})

	resp, err := loader.createRequest(DToken)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode != http.StatusOK, got %d", resp.StatusCode)
	}
}

func TestPrepareFile(t *testing.T) {

	fname := "file_"
	fSize := 1*1024*1024
	
	f := createFile(t, 4*1024*1024, TEST_DIR+string(filepath.Separator) + fname)
	defer f.Close()
	
	stat, _ := f.Stat()
	if stat.Size() == int64(fSize) {
		t.Fatal("stat.Size() == int64(fSize)")
	}
	
	l := NewLoader(CnfLoader{})
	if err := l.prepareFile(f, int64(fSize)); err != nil {
		t.Fatal(err)
	}

	newStat, _ := f.Stat() 
	if newStat.Size() != int64(fSize) {
		t.Fatal("newStat.Size() != int64(fSize)")
	}
}

func createFile(t *testing.T, size int64, fileName string) *os.File {
	
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, f)
	return f
}

func TestSaveFile(t *testing.T) {

	enc, err := ecnryption.NewEncryptor(ecnryption.CnfEncrypter{
		Key:       Key,
		Ecryption: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	loader := NewLoader(CnfLoader{
		Log:       logger,
		Addr:      AddrPath,
		BasePath:  TEST_DIR,
		Token:     Token,
		Encryptor: enc,
	})

	resp, err := loader.createRequest(DToken)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode != http.StatusOK, got %d", resp.StatusCode)
	}

	f := createFile(t,  4*1024*1024, TEST_DIR + "get_fake_test_file")
	defer f.Close()
	fw := FileWriter{
		Ctx: context.TODO(),
		F: f,
		Log: logger,
		WithEnctyption: false,
	}

	counter := WriteCounter{}

	if err := loader.saveFile(&fw, resp.Body, &counter); err != nil {
		t.Fatal(err)
	}
}