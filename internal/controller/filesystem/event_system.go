package filesystem

import (
	"context"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

const UPDATE_MODE = 200

type ISnapshotUsecase interface {
	CreateSnapshot(mode int)
}

// eventSystem.
type eventSystem struct {
	ctx             context.Context
	log             *logrus.Logger
	watcher         *fsnotify.Watcher
	snapshotUsecase ISnapshotUsecase
}

// CnfEventSystem.
type CnfEventSystem struct {
	Ctx             context.Context
	Log             *logrus.Logger
	BasePath        string
	SnapshotUsecase ISnapshotUsecase
}

// NewEventSystem.
func NewEventSystem(cnf CnfEventSystem) (*eventSystem, error) {

	if err := os.MkdirAll(cnf.BasePath, 0777); err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(cnf.BasePath); err != nil {
		return nil, err
	}

	return &eventSystem{
		ctx:             cnf.Ctx,
		log:             cnf.Log,
		snapshotUsecase: cnf.SnapshotUsecase,
		watcher:         watcher,
	}, nil
}

// Run.
func (e *eventSystem) Run() error {

	defer e.watcher.Close()

	makeSnapshotCh := make(chan struct{})

	go func() {
		ok := false
		for {
			select {
			case <-e.ctx.Done():
				return
			case <-time.Tick(1 * time.Second):
				if ok {
					e.snapshotUsecase.CreateSnapshot(UPDATE_MODE)
					ok = false
				}
			case <-makeSnapshotCh:
				ok = true
			}
		}
	}()

	for {
		select {
		case <-e.ctx.Done():
			return nil
		case event, ok := <-e.watcher.Events:
			if !ok {
				return nil
			}
			e.log.Infof("new event: %v", event)

			if !event.Has(fsnotify.Chmod) {
				makeSnapshotCh <- struct{}{}
			}
		case err, ok := <-e.watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}
