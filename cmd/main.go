package main

import (
	"context"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

func main() {
	e, err := NewEventSystem(CnfEventSystem{
		Ctx:      context.TODO(),
		Log:      logrus.New(),
		BasePath: "C:\\Users\\secrr\\Desktop\\нынешнее_программираммирование\\client\\test_dir",
	})
	if err != nil {
		panic(err)
	}
	if err := e.Run(); err != nil {
		panic(err)
	}
}

type eventSystem struct {
	ctx        context.Context
	log        *logrus.Logger
	basePath   string
	watcher    *fsnotify.Watcher
}

type CnfEventSystem struct {
	Ctx      context.Context
	Log      *logrus.Logger
	BasePath string
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
		ctx:        cnf.Ctx,
		log:        cnf.Log,
		basePath:   cnf.BasePath,
		watcher:    watcher,
	}, nil
}

func (e *eventSystem) Run() error {

	defer e.watcher.Close()

	makeSnapshotCh := make(chan struct{})

	go func() {
		ok := false
		for {
			select {
			case <-e.ctx.Done():
				return
			case <-time.Tick(1*time.Second):
				if ok {
					e.log.Info("run snapshot")
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

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				e.watcher.Add(event.Name)
				makeSnapshotCh <- struct{}{}
			}

			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				e.watcher.Remove(event.Name)
				makeSnapshotCh <- struct{}{}
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
