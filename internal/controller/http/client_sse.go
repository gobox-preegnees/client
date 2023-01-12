package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"

	"github.com/go-playground/validator/v10"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
)

// TODO: вынести в константы
const NEED_TO_DO = -1

type IdentifierUsecase interface {
	GetClientId() (clientID string)
	GetAddrSSE() (addr string)
}

type SnapshotUsecase interface {
	GetCurrentRequestId() (id int)
}

type ConsistencyUsecase interface {
	ApplyConsistency(entity.Consistency)
}

type Usecase interface {
	SnapshotUsecase
	IdentifierUsecase
	ConsistencyUsecase
}

// eventSSE.
type eventSSE struct {
	ctx       context.Context
	log       *logrus.Logger
	clientSSE *sse.Client
	usecase   Usecase
}

// CnfEventSSE.
type CnfEventSSE struct {
	Ctx      context.Context
	Log      *logrus.Logger
	JwtToken string
	Usecase  Usecase
}

// NewEventSSE.
func NewEventSSE(cnf CnfEventSSE) (*eventSSE, error) {

	client := sse.NewClient(cnf.Usecase.GetAddrSSE())
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if !client.Connected {
		return nil, fmt.Errorf("No connection")
	}

	return &eventSSE{
		ctx:       cnf.Ctx,
		log:       cnf.Log,
		clientSSE: client,
		usecase:   cnf.Usecase,
	}, nil
}

// Run.
func (e *eventSSE) Run() error {

	return e.clientSSE.SubscribeRaw(func(msg *sse.Event) {
		consistency := entity.Consistency{}
		if err := json.Unmarshal(msg.Data, &consistency); err != nil {
			e.log.Fatal(err)
		}

		v := validator.New()
		if err := v.Struct(consistency); err != nil {
			e.log.Fatal(err)
		}

		// Accept new chages
		if consistency.RequestId == NEED_TO_DO ||
			(consistency.Client == e.usecase.GetClientId() &&
				consistency.RequestId == e.usecase.GetCurrentRequestId()) {
			e.usecase.ApplyConsistency(consistency)
			e.log.Debugf("apply new consistency:%v", consistency)
			return
		}
		e.log.Debugf("ignore consistency:%v", consistency)
	})
}
