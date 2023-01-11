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

const NEED_TO_DO = -1

type IUsecase interface {
	// CurrentRequestId. Returns the current requestID
	GetCurrentRequestId() (id int)
	// Create a snapshot. Creating a snapshot of directory that you specified in the config
	CreateSnapshot()
	// MyClientId. Returns the clientID (name which need to create on web-sute)
	GetClientId() (clientID string)
	// MyClientId. Returns the address Server-Sent Events (SSE), example: http(s)://host:port/events?stream=<my_stream>
	GetAddrSSE() (addr string)
}

type eventSSE struct {
	ctx        context.Context
	log        *logrus.Logger
	clientSSE  *sse.Client
	usecase    IUsecase
}

type CnfEventSSE struct {
	Ctx      context.Context
	Log      *logrus.Logger
	JwtToken string
	Usecase  IUsecase
}

func NewEventSSE(cnf CnfEventSSE) (*eventSSE, error) {

	client := sse.NewClient(cnf.Usecase.GetAddrSSE())
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if !client.Connected {
		return nil, fmt.Errorf("No connection")
	}

	return &eventSSE{
		ctx:        cnf.Ctx,
		log:        cnf.Log,
		clientSSE:  client,
		usecase:    cnf.Usecase,
	}, nil
}

func (e *eventSSE) Run() error {

	return e.clientSSE.SubscribeRaw(func(msg *sse.Event) {
		// TODO: посмотеть, работает ли эта часть в цикле
		consistency := entity.Consistency{}
		if err := json.Unmarshal(msg.Data, &consistency); err != nil {
			e.log.Fatal(err)
		}
		v := validator.New()
		if err := v.Struct(consistency); err != nil {
			e.log.Fatal(err)
		}

		if consistency.RequestId == NEED_TO_DO {
			// Необходимо принять новые изменения в любом случае
			// Такой запрос означает, что но сервере обновились данные и нужно принять это обновление
		}
		if consistency.Client == e.usecase.GetClientId() {
			if consistency.RequestId == e.usecase.GetCurrentRequestId() {
				// Так как id совпадают, это значит, что это последний ответ, который мы ожидаем получить
			} else if consistency.RequestId < e.usecase.GetCurrentRequestId() {
				// Это принимать не нужно, так как ответ устарел и нужно дождаться последнего ответа
			} else {
				// panic("такого ответа быть не должно")
			}
		} else {
			// Это ответ не предназначался нашему клиенту, нужно пропустить
		}
	})
}