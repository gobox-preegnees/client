package usecase

import (
	"context"

	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

type IFileService interface {
	Lock(ctx context.Context, file entity.File) error
	Remove(ctx context.Context, file entity.File) error
	Raname(ctx context.Context, oldFile, newFile entity.File) error
	Upload(ctx context.Context, file entity.File, uToken string) error
	Download(ctx context.Context, file entity.File, dToken string) error
}

// TODO: использовать еще сервис снепшотов, чтобы при ошибке создавать снепшот ошибочный

// usecase.
type consistencyUsecase struct {
	log         *logrus.Logger
	fileService IFileService

	done            chan struct{}
	withoutSequence bool
}

// CnfConsistencyUsecase.
type CnfConsistencyUsecase struct {
	Log         *logrus.Logger
	FileService IFileService
}

// NewConsistencyUsecase.
func NewConsistencyUsecase(cnf CnfConsistencyUsecase) *consistencyUsecase {

	return &consistencyUsecase{
		log:             cnf.Log,
		fileService:     cnf.FileService,
		done:            make(chan struct{}, 1),
		withoutSequence: false,
	}
}

// WithoutSequence. Без последовательных событий. Новые консистенции могут применяться непоследовательно
func (u *consistencyUsecase) WithoutSequence() *consistencyUsecase {

	u.withoutSequence = true
	close(u.done)
	return u
}

// ApplyConsistency. Применяет изменения, последовательно, так как предполагается работа на мобильных устройствах
func (c *consistencyUsecase) ApplyConsistency(ctx context.Context, consistency entity.Consistency) {

	c.log.Debugf("New consistency: %v", consistency)

	if !c.withoutSequence {
		c.done <- struct{}{}
	}

	go func() {

		// c.sortNames()

		// TODO: тут можно запустить через errgroup
		if len(consistency.NeedToRemove) != 0 {
			for _, v := range consistency.NeedToRemove {
				c.fileService.Remove(ctx, v.File)
			}
		}

		if len(consistency.NeedToRename) != 0 {
			for _, v := range consistency.NeedToRename {
				c.fileService.Raname(ctx, v.OldFile, v.NewFile)
			}
		}

		if len(consistency.NeedToUpload) != 0 {
			for _, v := range consistency.NeedToUpload {
				c.fileService.Upload(ctx, v.File, v.Token)
			}
		}

		if len(consistency.NeedToDownload) != 0 {
			for _, v := range consistency.NeedToDownload {
				c.fileService.Download(ctx, v.File, v.Token)
			}
		}

		if !c.withoutSequence {
			<-c.done
		}
	}()
}

// TODO: это тоже вынести на уровень выше
// func (c consistencyUsecase) sortNames(files []entity.NeedToRename) {

// 	sort.SliceStable(files, func(i, j int) bool {
// 		len1 := len(strings.Split(files[i].OldFile.FileName, string(filepath.Separator)))
// 		len2 := len(strings.Split(files[j].OldFile.FileName, string(filepath.Separator)))
// 		return len1 > len2
// 	})
// }

// TODO: это нужно вынести ну уровень выше
// func (c consistencyUsecase) pathCorrection(consistency entity.Consistency) {
	
// 	if string(filepath.Separator) == "/" {
// 		return
// 	}

// 	for _, v := range consistency.NeedToRename {
// 		v.OldFile.FileName = strings.ReplaceAll(v.OldFile.FileName, "/", string(filepath.Separator))
// 		v.NewFile.FileName = strings.ReplaceAll(v.NewFile.FileName, "/", string(filepath.Separator))
// 	}

// 	for _, v :=  range consistency.NeedToDownload {
// 		v.File.FileName = strings.ReplaceAll(v.File.FileName, "/", string(filepath.Separator))
// 	}

// 	for _, v :=  range consistency.NeedToUpload {
// 		v.File.FileName = strings.ReplaceAll(v.File.FileName, "/", string(filepath.Separator))
// 	}

// 	for _, v :=  range consistency.NeedToLock {
// 		v.File.FileName = strings.ReplaceAll(v.File.FileName, "/", string(filepath.Separator))
// 	}

// 	for _, v :=  range consistency.NeedToRemove {
// 		v.File.FileName = strings.ReplaceAll(v.File.FileName, "/", string(filepath.Separator))
// 	}
// }
