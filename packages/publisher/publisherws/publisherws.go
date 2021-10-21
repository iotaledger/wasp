// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"net/http"
	"strings"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"nhooyr.io/websocket"
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
		log:      log.Named("PublisherWebSocket"),
		msgTypes: msgTypesMap,
	}
}

func (p *PublisherWebSocket) ServeHTTP(chainID *iscp.ChainID, w http.ResponseWriter, r *http.Request) error {
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

	cl := events.NewClosure(func(msgType string, parts []string) {
		if !p.msgTypes[msgType] {
			return
		}
		if len(parts) < 1 {
			return
		}
		if parts[0] != chainID.Base58() {
			return
		}

		select {
		case ch <- msgType + " " + strings.Join(parts, " "):
		default:
			p.log.Warnf("dropping websocket message for %s", r.RemoteAddr)
		}
	})
	publisher.Event.Attach(cl)
	defer publisher.Event.Detach(cl)

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
