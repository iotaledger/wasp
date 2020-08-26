package dashboard

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fr"
	"github.com/iotaledger/wasp/tools/wasp-client/config/tr"
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

var clients = map[string]*sync.Map{
	"fr": &sync.Map{},
	"fa": &sync.Map{},
	"tr": &sync.Map{},
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
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	e.Use(middleware.Recover())
	e.HideBanner = true
	e.Renderer = initRenderer()

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${time_rfc3339} ${level}")
	}
	e.Logger.SetLevel(log.INFO)

	e.GET("/", handleIndex)
	e.GET("/fairroulette", handleFR)
	e.GET("/fairauction", handleFA)
	e.GET("/tokenregistry", handleTR)
	e.GET("/tokenregistry/:color", handleTRQuery)
	e.GET("/ws/fr", handleWebSocket(fr.Config))
	e.GET("/ws/fa", handleWebSocket(fa.Config))
	e.GET("/ws/tr", handleWebSocket(tr.Config))

	done := startNanomsgForwarder(e.Logger)
	defer func() { done <- true }()

	e.Logger.Fatal(e.Start(listenAddr))
}

func addSCIfAvailable(availableSCs map[string]*config.SCConfig, sc *config.SCConfig) {
	scAddress := sc.TryAddress()
	if scAddress != nil {
		availableSCs[scAddress.String()] = sc
	}
}

func startNanomsgForwarder(logger echo.Logger) chan bool {
	done := make(chan bool)
	incomingStateMessages := make(chan []string)
	err := subscribe.Subscribe(config.WaspNanomsg(), incomingStateMessages, done, false, "state")
	check(err)
	logger.Infof("[Nanomsg] connected")

	availableSCs := make(map[string]*config.SCConfig)
	addSCIfAvailable(availableSCs, fr.Config)
	addSCIfAvailable(availableSCs, fa.Config)
	addSCIfAvailable(availableSCs, tr.Config)

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
					clients[sc.ShortName].Range(func(key interface{}, client interface{}) bool {
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

type Renderer map[string]*template.Template

func (t Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t[name].ExecuteTemplate(w, "base", data)
}

func initRenderer() Renderer {
	return Renderer{
		"index":         initIndexTemplate(),
		"fairroulette":  initFRTemplate(),
		"fairauction":   initFATemplate(),
		"tokenregistry": initTRTemplate(),
	}
}

func handleWebSocket(sc *config.SCConfig) func(c echo.Context) error {
	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			c.Logger().Infof("[WebSocket] opened for %s", c.Request().RemoteAddr)
			defer c.Logger().Infof("[WebSocket] closed for %s", c.Request().RemoteAddr)

			client := make(chan string)
			clients[sc.ShortName].Store(client, client)
			defer clients[sc.ShortName].Delete(client)

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
