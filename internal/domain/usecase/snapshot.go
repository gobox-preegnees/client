package usecase

import (
	"context"

	"github.com/sirupsen/logrus"
)

type ISnapshotService interface {
	CreateSnapshot(mode int)
	GetCurrentRequestId() (id int)
}

// snapshotUsecase.
type snapshotUsecase struct {
	log             *logrus.Logger
	snapshotService ISnapshotService
}

type CnfSnapshotUsecase struct {
	Ctx             context.Context
	Log             *logrus.Logger
	SnapshotUsecase ISnapshotService
}

func NewSnapshotUsecase(cnf CnfSnapshotUsecase) *snapshotUsecase {

	return &snapshotUsecase{
		log:             cnf.Log,
		snapshotService: cnf.SnapshotUsecase,
	}
}

// CreateSnapshot. Creates a new snapshot, will run until it creates and sends a snapshot.
// if will be new call then old proccess will stopped and will start a new.
func (s *snapshotUsecase) CreateSnapshot(mode int) {

	s.snapshotService.CreateSnapshot(mode)
}

// GetCurrentRequestId. Getting current requestId
func (s *snapshotUsecase) GetCurrentRequestId() (id int) {

	return s.snapshotService.GetCurrentRequestId()
}
