package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	lockedFileStorageDto "github.com/gobox-preegnees/gobox-client/internal/adapter/storage/locked_file"
	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"
	usecase "github.com/gobox-preegnees/gobox-client/internal/domain/usecase"

	"github.com/sirupsen/logrus"
)

var ErrCollision = errors.New("Err File Is Not Exist Or File Is Not Funded Or impossible to delete")
var ErrRenameFile = errors.New("Err Rename File")

var (
	errFindFile   = errors.New("err Find File")
	errPermission = errors.New(
		`The program wants to have full access to files and folders (777) in the sync folder.  
		Please fix the problem with file and folder permissions:)`,
	)
	errUnableToOpenFile = errors.New("unable to open file")
)

type ILockedFileStorage interface {
	// SaveLockFile. Сохраняются успешно заблокированные файлы
	SaveLockFile(req lockedFileStorageDto.CreateOneLockedFileDTO)
	// SaveUnlockFile. Сохраняются файлы, которые не получилось заблокировать
	SaveUnlockFile(req lockedFileStorageDto.CreateOneUnlockedFileDTO)
	// UnlockFile. Разблокировка файла
	UnlockFile(ctx context.Context, file entity.File)
	// FindAllLockedFiles() []entity.File
	// FindOneLockedFile(ctx context.Context, file entity.File) bool
}

type ILoader interface {
	Download(ctx context.Context, file entity.File, dToken string)
	Upload()
}

// fileService.
type fileService struct {
	log               *logrus.Logger
	lockedFileStorage ILockedFileStorage
	basePath          string
	encryptionKey     string
	loader            ILoader
}

// CnfFileService.
type CnfFileService struct {
	Log               *logrus.Logger
	LockedFileStorage ILockedFileStorage
	BasePath          string
	EncryptionKey     string
	Loader            ILoader
}

// NewFileService.
func NewFileService(cnf CnfFileService) *fileService {

	return &fileService{
		log:               cnf.Log,
		lockedFileStorage: cnf.LockedFileStorage,
		basePath:          cnf.BasePath,
		encryptionKey:     cnf.EncryptionKey,
		loader:            cnf.Loader,
	}
}

// Download implements usecase.IFileService.
func (f fileService) Download(ctx context.Context, file entity.File, dToken string) error { // TODO: file тут нужен для удобства, на уровне выше бинформация будет браться из токена

	if err := f.findFileWrap(file); err != nil {
		return err
	}

	// TODO: возможно сделать обработку ошибок
	f.loader.Download(ctx, file, dToken)
	return nil
}

// Upload implements usecase.IFileService.
func (f *fileService) Upload(ctx context.Context, file entity.File, uToken string) error {

	if err := f.findFileWrap(file); err != nil {
		return err
	}

	// TODO: возможно сделать обработку ошибок
	f.loader.Download(ctx, file, uToken)
	return nil
}

// Remove implements usecase.IFileService.
func (f *fileService) Remove(ctx context.Context, file entity.File) error {

	if err := f.findFileWrap(file); err != nil {
		return err
	}

	// Разблокировка (если он заблокирован), чтобы можно было удалить
	f.lockedFileStorage.UnlockFile(ctx, file)

	err := os.Remove(file.FileName)
	if err != nil {
		f.log.Error(err)
		// Так как если сейчас не получится переименовать,
		// то потом все файлы по новому пути будут отбрасваться
		return ErrRenameFile
	}
	f.log.Debugf("successfully remove file:%v", file)
	return nil
}

// Raname implements usecase.IFileService
func (f *fileService) Raname(ctx context.Context, oldFile, newFile entity.File) error {

	if err := f.findFileWrap(oldFile); err != nil {
		return err
	}

	// Разблокировка (если он заблокирован), чтобы можно было переименовать
	f.lockedFileStorage.UnlockFile(ctx, oldFile)

	err := os.Rename(oldFile.FileName, newFile.FileName)
	if err != nil {
		f.log.Error(err)
		// Так как если сейчас не получится переименовать,
		// то потом все файлы по новому пути будут отбрасваться
		return ErrRenameFile
	}
	f.log.Debugf("successfully rename file:%v, new file:%s", oldFile, newFile)
	return nil
}

// Lock implements usecase.IFileService. // TODO: еще может добавить метод для разблокировки файла
func (f *fileService) Lock(ctx context.Context, file entity.File) error {

	of, err := f.openFile(file)
	// TODO: проверять эту ошибку, когда потребуется записать в файл что то
	// if errors.Is(err, os.ErrClosed) {}
	// if errors.Is(err, os.ErrInvalid) {}
	// if errors.Is(err, os.ErrExist) {}

	// TODO: что то с этим обязательно сделать
	if errors.Is(err, errUnableToOpenFile) {
		f.log.Warn(err)
		f.lockedFileStorage.SaveUnlockFile(lockedFileStorageDto.CreateOneUnlockedFileDTO{
			Ctx:  ctx,
			File: file,
		})
		return nil
	} else if errors.Is(err, os.ErrNotExist) || errors.Is(err, errFindFile) {
		f.log.Warn(err)
		return ErrCollision
	} else if errors.Is(err, os.ErrPermission) {
		// TODO: сделать какую нибудь систему ответа на сервер, при неизвестных ошибках
		f.log.Fatal(errPermission)
	} else if err != nil {
		f.log.Error(err)
		return err
	}

	f.lockedFileStorage.SaveLockFile(lockedFileStorageDto.CreateOneLockedFileDTO{
		Ctx:  ctx,
		File: file,
		F:    of,
	})
	f.log.Debugf("successfully created locked file:%", file)
	return nil
}

// openFile. Открывает файл и таким обазом блокирует
func (f fileService) openFile(file entity.File) (*os.File, error) {

	err := f.findFile(file)
	if err != nil {
		return nil, err
	}

	of, err := os.OpenFile(file.FileName, os.O_RDWR, 0777)
	if err != nil {
		return nil, errUnableToOpenFile
	}
	f.log.Debugf("successfully locked file:%", file)
	return of, nil
}

// findFileWrap. Переиспользование обработки ошибок при вызове метода findFile
func (f fileService) findFileWrap(file entity.File) error {

	err := f.findFile(file)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, errFindFile) {
		f.log.Warn(err)
		return ErrCollision
	} else if err != nil {
		f.log.Error(err)
		return err
	}
	return nil
}

// findFile. Пытается напйти точно такой же файл (не только по имени)
func (f fileService) findFile(file entity.File) error {

	fi, err := os.Stat(file.FileName)
	if err != nil {
		return err
	}

	modTime := fi.ModTime().UTC().Unix()
	sizeFile := fi.Size()

	if modTime != file.ModTime {
		return fmt.Errorf("err:%w, expected modtime:%d != got:%d", errFindFile, file.ModTime, modTime)
	}

	if fi.IsDir() {
		if file.HashSum != "" || file.SizeFile != 0 {
			return fmt.Errorf("err:%w, expected file != got folder", errFindFile)
		}
		return nil
	} else {
		if sizeFile != file.SizeFile {
			return fmt.Errorf("err:%w, expected sizeFile:%d != got:%d", errFindFile, file.SizeFile, sizeFile)
		}
	}
	f.log.Debugf("successfully funded file:%", file)
	return nil
}

var _ usecase.IFileService = (*fileService)(nil)
