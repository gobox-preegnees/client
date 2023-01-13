package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	consts "github.com/gobox-preegnees/gobox-client/internal/consts"
	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"
	IClientSSEMock "github.com/gobox-preegnees/gobox-client/internal/mocks/controller/http/client_sse/ISnapshotUsecase"

	"github.com/golang/mock/gomock"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
)

const addr = "127.0.0.1:6655"
const streamId = "1"
const clientId = "1"
const requestId = 1

func sseServerMock(t *testing.T, newData chan entity.Consistency) {
	server := sse.New()
	server.CreateStream(streamId)

	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.ServeHTTP)

	t.Log(addr)
	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		t.Log("running accept message")
		for d := range newData {
			data, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			server.Publish(streamId, &sse.Event{
				Data: []byte(data),
			})
		}
		s.Shutdown(context.TODO())
	}()

	go func() {
		t.Log("running server")
		err := s.ListenAndServe()
		if err != nil {
			t.Fatal(err)
		}
	}()
}

func TestClientSSE(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	data := []struct {
		requestId int
		client    string
		apply     bool
	}{
		{
			requestId: consts.NEED_TO_DO,
			client:    clientId,
			apply:     true,
		},
		{
			requestId: requestId,
			client:    "error",
			apply:     false,
		},
		{
			requestId: requestId,
			client:    clientId,
			apply:     false,
		},
		{
			requestId: -100,
			client:    clientId,
			apply:     false,
		},
	}

	wantApplied := 2
	currentApplied := 0

	consistencyUsecaseMock := IClientSSEMock.NewMockIConsistencyUsecase(controller)
	consistencyUsecaseMock.EXPECT().ApplyConsistency(gomock.Any()).AnyTimes().Do(func(c entity.Consistency) { currentApplied++ })

	snapshotUsecaseMock := IClientSSEMock.NewMockISnapshotUsecase(controller)
	snapshotUsecaseMock.EXPECT().GetCurrentRequestId().Return(requestId).AnyTimes()

	newData := make(chan entity.Consistency)
	sseServerMock(t, newData)
	time.Sleep(1 * time.Second)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	clSSE := NewClientSSE(CnfClientSSE{
		Ctx:                context.TODO(),
		Log:                logger,
		ConsistencyUsecase: consistencyUsecaseMock,
		SnapshotUsecase:    snapshotUsecaseMock,
		AddrSSE:            "http://" + addr + "/events",
		StreamId:           streamId,
		ClientId:           clientId,
	})

	go func() {
		if err := clSSE.Run(); err != nil {
			t.Fatal(err)
		}
	}()

	for _, v := range data {
		t.Run("testing run", func(t *testing.T) {
			consistency := entity.Consistency{
				RequestId:      v.requestId,
				Client:         v.client,
				NeedToRemove:   []entity.NeedToRemove{},
				NeedToRename:   []entity.NeedToRename{},
				NeedToUpload:   []entity.NeedToUpload{},
				NeedToDownload: []entity.NeedToDownload{},
			}
			newData <- consistency
		})
	}

	time.Sleep(2*time.Second)
	close(newData)

	if wantApplied != currentApplied {
		t.Fatalf("wantApplied{%d} != currentApplied{%d}", wantApplied, currentApplied)
	}
}
