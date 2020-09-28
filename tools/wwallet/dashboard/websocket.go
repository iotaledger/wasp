package dashboard

import (
	"strings"
	"sync"

	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

var wsClients = map[string]*sync.Map{}

func addWebSocketTab(e *echo.Echo, sc *sc.Config) {
	wsClients[sc.ShortName] = &sync.Map{}
	e.GET("/ws/"+sc.ShortName, handleWebSocket(sc))
}

func startNanomsgForwarder(logger echo.Logger, availableSCs map[string]*sc.Config) chan bool {
	done := make(chan bool)
	incomingStateMessages := make(chan []string)
	err := subscribe.Subscribe(config.WaspNanomsg(), incomingStateMessages, done, false, "state")
	check(err)
	logger.Infof("[Nanomsg] connected")

	go func() {
		for {
			select {
			case msg := <-incomingStateMessages:
				scAddress := msg[1]
				sc, ok := availableSCs[scAddress]
				if !ok {
					continue
				}
				{
					msg := strings.Join(msg, " ")
					logger.Infof("[Nanomsg] got message %s", msg)
					wsClients[sc.ShortName].Range(func(key interface{}, client interface{}) bool {
						if client, ok := client.(chan string); ok {
							client <- msg
						}
						return true
					})
				}
			case <-done:
				logger.Infof("[Nanomsg] closing connection...")
				break
			}
		}
	}()
	return done
}

func handleWebSocket(sc *sc.Config) func(c echo.Context) error {
	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			c.Logger().Infof("[WebSocket] opened for %s", c.Request().RemoteAddr)
			defer c.Logger().Infof("[WebSocket] closed for %s", c.Request().RemoteAddr)

			client := make(chan string)
			wsClients[sc.ShortName].Store(client, client)
			defer wsClients[sc.ShortName].Delete(client)

			for {
				msg := <-client
				_, err := ws.Write([]byte(msg))
				if err != nil {
					break
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
