package service

import (
	cacheDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/storage/cache"
	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"
	"github.com/sirupsen/logrus"
)

type ICache interface {
	Get(cacheDTO.GetFileDTO) entity.File
	Put(cacheDTO.PutFileDTO)
	GetAll() (string, []entity.File)
}

type snapshotService struct {
	log      *logrus.Logger
	cache    ICache
	basePath string
}

func NewCnapshotService() *snapshotService {

	return &snapshotService{}
}

func (s *snapshotService) CreateSnapshots() error {

	return s.create(s.basePath)
}

func (s *snapshotService) create(path string) error {


	return nil
}
