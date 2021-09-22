// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"net/http"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"nhooyr.io/websocket"
)

type PublisherWebSocket struct {
	clients sync.Map
	closure *events.Closure
}

func New() *PublisherWebSocket {
	return &PublisherWebSocket{}
}

func (p *PublisherWebSocket) ServeHTTP(chainID *iscp.ChainID, w http.ResponseWriter, r *http.Request) error {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusInternalError, "something went wrong")

	v, _ := p.clients.LoadOrStore(chainID.Base58(), &sync.Map{})
	chainWsClients := v.(*sync.Map)

	clientCh := make(chan string)
	chainWsClients.Store(clientCh, clientCh)
	defer chainWsClients.Delete(clientCh)

	ctx := c.CloseRead(r.Context())
	for {
		msg := <-clientCh
		err = c.Write(ctx, websocket.MessageBinary, []byte(msg))
		if err != nil {
			c.Close(websocket.StatusInternalError, err.Error())
			break
		}
	}
	return nil
}

func (p *PublisherWebSocket) Start(msgTypes ...string) {
	if p.closure != nil {
		panic("Start called twice")
	}

	msgTypesMap := make(map[string]bool)
	for _, t := range msgTypes {
		msgTypesMap[t] = true
	}

	p.closure = events.NewClosure(func(msgType string, parts []string) {
		if msgTypesMap[msgType] {
			if len(parts) < 1 {
				return
			}
			chainID := parts[0]

			v, ok := p.clients.Load(chainID)

			if !ok {
				return
			}
			chainWsClients := v.(*sync.Map)

			msg := msgType + " " + strings.Join(parts, " ")
			chainWsClients.Range(func(key interface{}, clientCh interface{}) bool {
				clientCh.(chan string) <- msg
				return true
			})
		}
	})
	publisher.Event.Attach(p.closure)
}

func (p *PublisherWebSocket) Stop() {
	publisher.Event.Detach(p.closure)
	p.closure = nil
}
