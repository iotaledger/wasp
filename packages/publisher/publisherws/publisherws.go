// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"net/http"
	"strings"

	"nhooyr.io/websocket"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher"
)

type PublisherWebSocket struct {
	log       *logger.Logger
	publisher *publisher.Publisher
	msgTypes  map[string]bool
}

func New(log *logger.Logger, publisher *publisher.Publisher, msgTypes []string) *PublisherWebSocket {
	msgTypesMap := make(map[string]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	return &PublisherWebSocket{
		log:       log.Named("PublisherWebSocket"),
		publisher: publisher,
		msgTypes:  msgTypesMap,
	}
}

func (p *PublisherWebSocket) ServeHTTP(chainID isc.ChainID, w http.ResponseWriter, r *http.Request) error {
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

	ch := make(chan string, 10)

	cl := event.NewClosure(func(ev *publisher.PublishedEvent) {
		if !p.msgTypes[ev.MsgType] {
			return
		}
		if len(ev.MsgType) < 1 {
			return
		}
		if ev.ChainID != chainID {
			return
		}

		select {
		case ch <- ev.MsgType + " " + strings.Join(ev.Parts, " "):
		default:
			p.log.Warnf("dropping websocket message for %s", r.RemoteAddr)
		}
	})

	if p.publisher != nil {
		p.publisher.Events.Published.Hook(cl)
		defer p.publisher.Events.Published.Detach(cl)
	}

	for {
		msg := <-ch
		err := c.Write(ctx, websocket.MessageText, []byte(msg))
		if err != nil {
			c.Close(websocket.StatusInternalError, err.Error())
			break
		}
	}

	return nil
}
