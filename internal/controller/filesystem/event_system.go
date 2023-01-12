package filesystem

/*
* The "filesystem" package is needed to intercept a new file system event 
* and call the snapshot creation function.
*/

import (
	"context"
	"os"
	"time"

	consts "github.com/gobox-preegnees/gobox-client/internal/consts"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -destination=../../mocks/controller/filesystem/event_system/ISnapshotUsecase/ISnapshotUsecase.go -source=event_system.go
type ISnapshotUsecase interface {
	// Creating snapshot using some mode (see github.com/gobox-preegnees/gobox-client/internal/consts)
	CreateSnapshot(mode int)
}

// eventSystem.
type eventSystem struct {
	ctx             context.Context
	log             *logrus.Logger
	watcher         *fsnotify.Watcher
	snapshotUsecase ISnapshotUsecase
}

// CnfEventSystem. All fields are required
type CnfEventSystem struct {
	Ctx context.Context
	Log *logrus.Logger
	// BasePath. This path should have been specified in the config
	BasePath        string
	SnapshotUsecase ISnapshotUsecase
}

// NewEventSystem. Create a new event-system, returns errpr mkdir or fsnotify
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

	cnf.Log.Debug("Created new event-system instance")
	return &eventSystem{
		ctx:             cnf.Ctx,
		log:             cnf.Log,
		snapshotUsecase: cnf.SnapshotUsecase,
		watcher:         watcher,
	}, nil
}

// Run. Returns the error watcher.Errors
func (e *eventSystem) Run() error {

	defer e.watcher.Close()

	makeSnapshotCh := make(chan struct{})

	go func() {
		// file directory change indicator
		ok := false
		for {
			select {
			case <-e.ctx.Done():
				return
			case <-time.Tick(1 * time.Second):
				// so as not to force a large number of times to create a snapshot
				if ok {
					e.snapshotUsecase.CreateSnapshot(consts.UPDATE_MODE)
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
			e.watcher.Close()
			close(makeSnapshotCh)
			return nil
		case event, ok := <-e.watcher.Events:
			if !ok {
				return nil
			}
			e.log.Debugf("new event: %v", event)

			// This app is not supported Chmod
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
