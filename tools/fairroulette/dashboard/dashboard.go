package dashboard

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/client"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Debug = true
	e.HideBanner = true
	e.Renderer = renderer

	e.GET("/", index)
	e.GET("/ws", ws)

	fmt.Printf("Serving dashboard on %s\n", listenAddr)
	e.Logger.Fatal(e.Start(listenAddr))
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
		messages := make(chan []string)
		done := make(chan bool)
		defer func() { done <- true }()
		err := subscribe.Subscribe(config.WaspNanomsg(), messages, done, false, "vmmsg")
		if err != nil {
			c.Logger().Error(err)
			return
		}
		for {
			select {
			case msg := <-messages:
				progHash := msg[1]
				if progHash != fairroulette.ProgramHash {
					continue
				}
				_, err := ws.Write([]byte(strings.Join(msg[2:], " ")))
				if err != nil {
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
