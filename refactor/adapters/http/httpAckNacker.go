// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// httpAckNacker implements the AckNacker interface
type httpAckNacker struct {
	Chresp chan<- MsgRes // A channel dedicated to send back a response
}

// Ack implements the core.AckNacker interface
func (an httpAckNacker) Ack(p *core.Packet) error {
	if an.Chresp == nil {
		return nil
	}
	defer close(an.Chresp)

	if p == nil {
		an.Chresp <- MsgRes{StatusCode: http.StatusOK}
		return nil
	}

	raw, err := json.Marshal(*p)
	if err != nil {
		return errors.New(ErrInvalidStructure, err)
	}

	select {
	case an.Chresp <- MsgRes{
		StatusCode: http.StatusOK,
		Content:    raw,
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(ErrWrongBehavior, "No response was given to the acknacker")
	}
}

// Nack implements the core.AckNacker interface
func (an httpAckNacker) Nack() error {
	if an.Chresp == nil {
		return nil
	}
	defer close(an.Chresp)

	select {
	case an.Chresp <- MsgRes{
		StatusCode: http.StatusNotFound,
		Content:    []byte(`{"message":"Not in charge of the associated device"}`),
	}:
	case <-time.After(time.Millisecond * 50):
		return errors.New(ErrWrongBehavior, "No response was given to the acknacker")
	}
	return nil
}
