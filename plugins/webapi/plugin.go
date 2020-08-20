package webapi

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/labstack/echo"
)

// PluginName is the name of the web API plugin.
const PluginName = "WebAPI"

var (
	// Server is the web API server.
	Server = echo.New()

	log *logger.Logger

	initWG sync.WaitGroup
)

func Init() *node.Plugin {
	Plugin := node.NewPlugin(PluginName, node.Enabled, configure, run)
	initWG.Add(1)
	return Plugin
}

func WaitUntilIsUp() {
	initWG.Wait()
}

func configure(*node.Plugin) {
	log = logger.NewLogger(PluginName)
	dkgapi.InitLogger()
	admapi.InitLogger()

	Server.HideBanner = true
	Server.HidePort = true
	addEndpoints()
}

func run(_ *node.Plugin) {
	log.Infof("Starting %s ...", PluginName)
	if err := daemon.BackgroundWorker("WebAPI Server", worker, parameters.PriorityWebAPI); err != nil {
		log.Errorf("Error starting as daemon: %s", err)
	}

	initWG.Done()
}

func worker(shutdownSignal <-chan struct{}) {
	defer log.Infof("Stopping %s ... done", PluginName)

	stopped := make(chan struct{})
	bindAddr := parameters.GetString(parameters.WebAPIBindAddress)
	go func() {
		log.Infof("%s started, bind-address=%s", PluginName, bindAddr)
		if err := Server.Start(bindAddr); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("Error serving: %s", err)
			}
			close(stopped)
		}
	}()

	// stop if we are shutting down or the server could not be started
	select {
	case <-shutdownSignal:
	case <-stopped:
	}

	log.Infof("Stopping %s ...", PluginName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Server.Shutdown(ctx); err != nil {
		log.Errorf("Error stopping: %s", err)
	}
}
