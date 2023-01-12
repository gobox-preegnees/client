package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	consts "github.com/gobox-preegnees/gobox-client/internal/consts"
	ISnapshotUsecaseMock "github.com/gobox-preegnees/gobox-client/internal/mocks/controller/filesystem/event_system/ISnapshotUsecase"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

const TEST_DIR = "TEST_DIR"

func TestEventSystem(t *testing.T) {

	os.MkdirAll(TEST_DIR, 0777)
	defer os.RemoveAll(TEST_DIR)

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	snapshotUsecaseMock := ISnapshotUsecaseMock.NewMockISnapshotUsecase(mockController)
	snapshotUsecaseMock.
		EXPECT().
		CreateSnapshot(consts.UPDATE_MODE).
		Times(1).
		Do(func(i int) {
			t.Log(i)
			cancel()
		})

	eventSystem, err := NewEventSystem(CnfEventSystem{
		Ctx:             ctx,
		Log:             logger,
		BasePath:        TEST_DIR,
		SnapshotUsecase: snapshotUsecaseMock,
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		f, err := os.Create(filepath.Join(TEST_DIR, "new_file.txt"))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}()

	if err := eventSystem.Run(); err != nil {
		t.Fatal(err)
	}
}
