// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package publisherws

import (
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"golang.org/x/net/websocket"
)

type PublisherWebSocket struct {
	clients sync.Map
	closure *events.Closure
}

func New() *PublisherWebSocket {
	return &PublisherWebSocket{}
}

func (p *PublisherWebSocket) GetHandler(chainID *iscp.ChainID) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		p.handleClient(ws, chainID)
	})
}

func (p *PublisherWebSocket) handleClient(ws *websocket.Conn, chainID *iscp.ChainID) {
	defer ws.Close()

	v, _ := p.clients.LoadOrStore(chainID.Base58(), &sync.Map{})
	chainWsClients := v.(*sync.Map)

	clientCh := make(chan string)
	chainWsClients.Store(clientCh, clientCh)
	defer chainWsClients.Delete(clientCh)

	for {
		msg := <-clientCh
		_, err := ws.Write([]byte(msg))
		if err != nil {
			break
		}
	}
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
