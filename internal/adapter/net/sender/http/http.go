package http

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	senderDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/net/sender"
)

var ErrNotOk = errors.New("not ok")

type sender struct {
	Addr      string
	AuthToken string
}

type CnfSender struct {
}

func NewSender() *sender {

	return &sender{}
}

func (s *sender) SendSnapshot(in senderDTO.SendSnapshotDTO) error {

	req, err := http.NewRequest(http.MethodPost, s.Addr, bytes.NewBuffer(in.Snapshot))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	req.Close = true

	cli := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("%w, code: %d", ErrNotOk, resp.StatusCode)
	}
}
