package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	cacheDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/storage/cache"
	senderDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/net/sender"
	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

const ERROR_MODE = 300
const ON_START_MODE = 100

var ErrReadDir = errors.New("err reading directory")

// TODO: сделать вывод через хуки logrus
// TODO: проверить каналы на закрытость

type ICache interface {
	Get(cacheDTO.GetFileDTO) (entity.File, bool)
	Put(cacheDTO.PutFileDTO)
	GetAll() (string, []entity.File)
}

type ISenderAdapter interface {
	SendSnapshot(senderDTO.SendSnapshotDTO) error
}

type snapshotService struct {
	log      *logrus.Logger
	
	cache    ICache
	sender   ISenderAdapter
	
	basePath string

	client string

	// req id
	currentId int
	tmpId     int

	ctxExternal    context.Context
	cancelInternal context.CancelFunc

	// identifier which identifies about running of snapshot
	running bool
	// mode snapshot: error, on_start, update
	mode int
}

func NewCnapshotService() *snapshotService {

	return &snapshotService{}
}

// CreateSnapshot. Creates a new snapshot, will run until it creates and sends a snapshot.
// if will be new call then old proccess will stopped and will start a new.
func (s *snapshotService) CreateSnapshot(mode int) {

	// TODO: возможно тут лучше использовать ctx.Err() -> context.Canceled
	if s.running {
		s.mode = mode
		s.cancelInternal()
		return
	}

	go func() {
		ctxInternal, cancelInternal := context.WithCancel(s.ctxExternal)
		s.cancelInternal = cancelInternal

		s.running = true
		defer func() {
			s.running = false
		}()

		for {
			select {
			case <-s.ctxExternal.Done():
				return
			default:
				var snapshot entity.Snapshot
				var err error
				for {
					snapshot, err = s.createsSnapshot(ctxInternal)
					if err == nil {
						break
					}
					time.Sleep(1 * time.Second)
				}

				if err != nil && s.mode != ON_START_MODE {
					s.mode = ERROR_MODE
				}

				snapshot.RequesID = s.nextRequestId()
				snapshot.Mode = s.mode
				snapshot.Client = s.client

				err = s.sendSnapshot(snapshot)
				if err == nil {
					s.commitRequestId()
					if errors.Is(ctxInternal.Err(), context.Canceled) {
						continue
					}
					break
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (s *snapshotService) createsSnapshot(ctx context.Context) (entity.Snapshot, error) {

	fileCh := make(chan entity.File)
	errCh := make(chan error)
	done := make(chan struct{})
	go func() {
		s.create(ctx, s.basePath, errCh, fileCh)
		done <- struct{}{}
	}()

	closeAll := func() {
		close(done)
		close(errCh)
		close(fileCh)
	}

	files := make([]entity.File, 0)

	for {
		select {
		case <-s.ctxExternal.Done():
			// closeAll
			return entity.Snapshot{}, nil
		case <-ctx.Done():
			// closeAll
			return entity.Snapshot{}, context.Canceled
		case err := <-errCh:
			closeAll()
			return entity.Snapshot{}, err
		case file := <-fileCh:
			files = append(files, file)
		case <-done:
			closeAll()
			return entity.Snapshot{
				Files: files,
			}, nil
		}
	}
}

func (s *snapshotService) create(ctx context.Context, basePath string, errCh chan<- error, fileCh chan<- entity.File) {

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		// TODO: обработать сверху
		errCh <- fmt.Errorf("%w dir:%s", ErrReadDir, basePath)
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			// TODO: обработать сверху
			errCh <- context.Canceled
		default:
			modTime := file.ModTime().UTC().Unix()
			fileName := filepath.ToSlash(filepath.Join(basePath, file.Name()))
			if file.IsDir() {
				fileCh <- entity.File{
					FileName: fileName,
					ModTime:  modTime,
					HashSum:  "",
					SizeFile: 0,
				}
				s.create(ctx, fileName, errCh, fileCh)
			} else {
				fromCache, ok := s.cache.Get(cacheDTO.GetFileDTO{
					FileName: fileName,
					ModTime:  modTime,
				})
				if ok {
					fileCh <- fromCache
				} else {
					f, err := os.OpenFile(fileName, os.O_RDWR, 0777)
					if err != nil {
						errCh <- err
						return
					}
					defer f.Close()

					h := sha256.New()
					if _, err := io.Copy(h, f); err != nil {
						errCh <- err
					}
					hashSum := h.Sum(nil)
					sizeFile := file.Size()
					s.cache.Put(cacheDTO.PutFileDTO{
						FileName: fileName,
						ModTime:  modTime,
						HashSum:  string(hashSum),
						SizeFile: sizeFile,
					})
					fileCh <- entity.File{
						FileName: fileName,
						HashSum:  string(hashSum),
						SizeFile: sizeFile,
						ModTime:  modTime,
					}
				}
			}
		}
	}
}

func (s *snapshotService) nextRequestId() int {

	s.tmpId = s.currentId + 1
	return s.tmpId
}

func (s *snapshotService) commitRequestId() {

	s.currentId = s.tmpId
}

func (s *snapshotService) sendSnapshot(snapshot entity.Snapshot) error {

	s.sender.SendSnapshot(senderDTO.SendSnapshotDTO{
		
	})
		panic("")
}
