package http

/*
* The "http" package is needed to getting a new consistency
* and call the applying function.
 */

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	consts "github.com/gobox-preegnees/gobox-client/internal/consts"
	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"

	"github.com/go-playground/validator/v10"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -destination=../../mocks/controller/http/client_sse/ISnapshotUsecase/ISnapshotUsecase.go -source=client_sse.go
type ISnapshotUsecase interface {
	// Getting a new ID, since it changes when new snapshots are sent
	GetCurrentRequestId() (id int)
}

//go:generate mockgen -destination=../../mocks/controller/http/client_sse/IConsistencyUsecase/IConsistencyUsecase.go -source=client_sse.go
type IConsistencyUsecase interface {
	// Accepts the new consistency that came from the server
	ApplyConsistency(entity.Consistency)
}

// eventSSE.
type clientSSE struct {
	ctx                context.Context
	log                *logrus.Logger
	clientSSE          *sse.Client
	consistencyUsecase IConsistencyUsecase
	snapshotUsecase    ISnapshotUsecase
	clientId           string
	streamId           string
}

// CnfEventSSE.
type CnfClientSSE struct {
	Ctx                context.Context
	Log                *logrus.Logger
	ConsistencyUsecase IConsistencyUsecase
	SnapshotUsecase    ISnapshotUsecase
	StreamId           string
	// Example: http(s)://localhost:8080/events?stream=1
	AddrSSE  string
	ClientId string
}

// NewEventSSE.
func NewClientSSE(cnf CnfClientSSE) *clientSSE {

	client := sse.NewClient(cnf.AddrSSE)
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &clientSSE{
		ctx:                cnf.Ctx,
		log:                cnf.Log,
		streamId:             cnf.StreamId,
		clientSSE:          client,
		clientId:           cnf.ClientId,
		consistencyUsecase: cnf.ConsistencyUsecase,
		snapshotUsecase:    cnf.SnapshotUsecase,
	}
}

// Run.
func (e clientSSE) Run() error {

	err := e.clientSSE.Subscribe(e.streamId, func(msg *sse.Event) {
		consistency := entity.Consistency{}
		if err := json.Unmarshal(msg.Data, &consistency); err != nil {
			e.log.Fatal(err)
		}

		v := validator.New()
		if err := v.Struct(consistency); err != nil {
			e.log.Fatal(err)
		}

		// TODO: тут где то нужно отсортировать пути для переименования, либо сделать это на сервере
		// и поменять / на \, в зависимости от платформы!!!
		// и еще нужно где то расшифровать

		// Accept new chages if NEED_TO_DO is true or client and requestId is valid
		if consistency.RequestId == consts.NEED_TO_DO ||
			(consistency.Client == e.clientId &&
				consistency.RequestId == e.snapshotUsecase.GetCurrentRequestId()) {
			e.consistencyUsecase.ApplyConsistency(consistency)
			e.log.Debugf("apply new consistency:%v", consistency)
			return
		}
		e.log.Debugf("ignore consistency:%v", consistency)
	})
	if err != nil {
		return err
	}
	return nil
}
