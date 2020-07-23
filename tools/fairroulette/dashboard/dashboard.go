package dashboard

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/tools/fairroulette/client"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/websocket"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var clients sync.Map

func Cmd(args []string) {
	listenAddr := ":10000"
	if len(args) > 0 {
		if len(args) != 1 {
			fmt.Printf("Usage: %s dashboard [listen-address]\n", os.Args[0])
			os.Exit(1)
		}
		listenAddr = args[0]
	}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	e.Use(middleware.Recover())
	e.HideBanner = true
	e.Renderer = renderer

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${time_rfc3339} ${level}")
	}
	e.Logger.SetLevel(log.INFO)

	e.GET("/", index)
	e.GET("/ws", ws)

	done := startNanomsgForwarder(e.Logger)
	defer func() { done <- true }()

	e.Logger.Fatal(e.Start(listenAddr))
}

func startNanomsgForwarder(logger echo.Logger) chan bool {
	done := make(chan bool)
	incomingStateMessages := make(chan []string)
	err := subscribe.Subscribe(config.WaspNanomsg(), incomingStateMessages, done, false, "state")
	check(err)
	logger.Infof("[Nanomsg] connected")

	scAddress := config.GetSCAddress().String()

	go func() {
		for {
			select {
			case msg := <-incomingStateMessages:
				addr := msg[1]
				if addr != scAddress {
					continue
				}
				{
					msg := strings.Join(msg, " ")
					logger.Infof("[Nanomsg] got message %s", msg)
					clients.Range(func(key interface{}, client interface{}) bool {
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

func index(c echo.Context) error {
	status, err := client.FetchStatus()
	if err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(c.Request().Host)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "index", &IndexTemplateParams{
		Host:      host,
		SCAddress: config.GetSCAddress(),
		Status:    status,
	})
}

func ws(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		c.Logger().Infof("[WebSocket] opened for %s", c.Request().RemoteAddr)
		defer c.Logger().Infof("[WebSocket] closed for %s", c.Request().RemoteAddr)

		client := make(chan string)
		clients.Store(client, client)
		defer clients.Delete(client)

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
