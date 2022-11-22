// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"encoding/json"
	"net/http"

	"nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher"
)

type PublisherWebSocket struct {
	log      *logger.Logger
	msgTypes map[string]bool
}

func New(log *logger.Logger, msgTypes []string) *PublisherWebSocket {
	msgTypesMap := make(map[string]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	return &PublisherWebSocket{
		log:      log.Named("PublisherWebSocketJSON"),
		msgTypes: msgTypesMap,
	}
}

func (p *PublisherWebSocket) ServeHTTP(chainID *isc.ChainID, w http.ResponseWriter, r *http.Request) error {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // TODO: make accept origin configurable
	})
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusInternalError, "something went wrong")
	ctx := c.CloseRead(r.Context())

	p.log.Debugf("accepted websocket connection from %s", r.RemoteAddr)
	defer p.log.Debugf("closed websocket connection from %s", r.RemoteAddr)

	ch := make(chan *publisher.ChainEvent, 10)

	cl := events.NewClosure(func(event *publisher.ChainEvent) {
		if !p.msgTypes[event.MessageType] {
			return
		}

		if chainID.String() != event.ChainID {
			return
		}

		select {
		case ch <- event:
		default:
			p.log.Warnf("dropping websocket message for %s", r.RemoteAddr)
		}
	})
	publisher.Event.Hook(cl)
	defer publisher.Event.Detach(cl)

	for {
		msg := <-ch
		event, err := json.Marshal(msg)
		if err != nil {
			break
		}

		err = c.Write(ctx, websocket.MessageText, event)
		if err != nil {
			c.Close(websocket.StatusInternalError, err.Error())
			break
		}
	}

	return nil
}
