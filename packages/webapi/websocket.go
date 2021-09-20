package webapi

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"golang.org/x/net/websocket"
)

type webSocketAPI struct {
	wsClients sync.Map
}

func addWsEndpoint(e echoswagger.ApiGroup) *webSocketAPI {
	api := &webSocketAPI{
		wsClients: sync.Map{},
	}

	e.GET("/chain/:chainid/ws", api.handleWebSocket)

	return api
}

func (w *webSocketAPI) handleWebSocket(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		c.Logger().Infof("[WebSocket] opened for %s", c.Request().RemoteAddr)
		defer c.Logger().Infof("[WebSocket] closed for %s", c.Request().RemoteAddr)

		v, _ := w.wsClients.LoadOrStore(chainID.Base58(), &sync.Map{})
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
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func (w *webSocketAPI) startWsForwarder() {
	cl := events.NewClosure(func(msgType string, parts []string) {
		if msgType == "state" || msgType == "vmmsg" {
			if len(parts) < 1 {
				return
			}
			chainID := parts[0]

			v, ok := w.wsClients.Load(strings.Replace(chainID, "$/", "", -1))

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
	publisher.Event.Attach(cl)
}
