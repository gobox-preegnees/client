package usecase

import (
	"context"
	"time"

	"github.com/gobox-preegnees/gobox-client/internal/domain/entity"
	"github.com/sirupsen/logrus"
)

const ERROR_MODE = 300
const ON_START_MODE = 100

// TODO: сделать вывод через хуки logrus

type ISnapshotService interface {
	// CreateSnapshot. Create new snapshot of directory, maybe canceled
	CreateSnapshot(ctx context.Context) (entity.Snapshot, error)
	// SendSnapshot. Send snapshot to server
	SendSnapshot(entity.Snapshot) error
	// GetCurrentRequestId. Getting current requestId
	GetCurrentRequestId() (id int)
	// NextRequestId. Creating next requestId (1 -> 2)
	NextRequestId() (id int)
	// RollbackRequestId. Roll backing requestId (2 -> 1)
	RollbackRequestId()
	// CommitRequestId. Commit requestId, method GetCurrentRequestId will be returned new commited requestID
	CommitRequestId()
}

// snapshotUsecase.
type snapshotUsecase struct {
	ctxExternal     context.Context
	cancelInternal  context.CancelFunc
	log             *logrus.Logger
	basePath        string
	snapshotService ISnapshotService
	// identifier which identifies about running of snapshot
	running bool
	// mode snapshot: error, on_start, update
	mode int
}

type CnfSnapshotUsecase struct {
	Ctx             context.Context
	Log             *logrus.Logger
	BasePath        string
	SnapshotUsecase ISnapshotService
}

func NewSnapshotUsecase(cnf CnfSnapshotUsecase) *snapshotUsecase {

	return &snapshotUsecase{
		ctxExternal:     cnf.Ctx,
		log:             cnf.Log,
		basePath:        cnf.BasePath,
		snapshotService: cnf.SnapshotUsecase,
	}
}

// CreateSnapshot. Creates a new snapshot, will run until it creates and sends a snapshot.
// if will be new call then old proccess will stopped and will start a new.
func (s *snapshotUsecase) CreateSnapshot(mode int) {

	if s.running {
		s.cancelInternal()
		return
	}

	go func() {
		s.mode = mode
	
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
					snapshot, err = s.snapshotService.CreateSnapshot(ctxInternal)
					if err == nil {
						break
					}
					time.Sleep(1 * time.Second)
				}

				if err != nil && s.mode != ON_START_MODE {
					s.mode = ERROR_MODE
				}

				snapshot.RequesID = s.snapshotService.NextRequestId()
				snapshot.Mode = s.mode
				if err := s.snapshotService.SendSnapshot(snapshot); err == nil {
					s.snapshotService.CommitRequestId()
					break
				} else {
					s.snapshotService.RollbackRequestId()
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

// GetCurrentRequestId. Getting current requestId
func (s *snapshotUsecase) GetCurrentRequestId() (id int) {

	return s.snapshotService.GetCurrentRequestId()
}
