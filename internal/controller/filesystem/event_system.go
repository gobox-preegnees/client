package filesystem

import (
	"context"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type ISnapshotUsecase interface {
	CreateSnapshot()
}

type eventSystem struct {
	ctx             context.Context
	log             *logrus.Logger
	basePath        string
	watcher         *fsnotify.Watcher
	snapshotUsecase ISnapshotUsecase
	snapshotCh      chan struct{}
}

type CnfEventSystem struct {
	Ctx             context.Context
	Log             *logrus.Logger
	BasePath        string
	SnapshotUsecase ISnapshotUsecase
}

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

	// TODO: Занести все файлы перед началом

	return &eventSystem{
		ctx:             cnf.Ctx,
		log:             cnf.Log,
		basePath:        cnf.BasePath,
		snapshotUsecase: cnf.SnapshotUsecase,
		watcher:         watcher,
		snapshotCh:      make(chan struct{}),
	}, nil
}

func (e *eventSystem) Run() error {

	defer e.watcher.Close()

	for {
		select {
		case <-e.ctx.Done():
			return nil
		case event, ok := <-e.watcher.Events:
			if !ok {
				return nil
			}
			e.log.Infof("new event: %v", event)

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				e.watcher.Add(event.Name)
				e.snapshotCh <- struct{}{}
			}

			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				e.watcher.Remove(event.Name)
				e.snapshotCh <- struct{}{}
			}

			// fsnotify.Chmod не предусмотрен, будет игнорироваться

		case err, ok := <-e.watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}

// TODO: взять из main
func (e *eventSystem) createSnapshot() {

	for {
		select {
		case <-e.snapshotCh:
			<- time.After(500 * time.Millisecond)
			e.snapshotUsecase.CreateSnapshot()
			e.log.Info("run snapshot")
		}
	}
}
