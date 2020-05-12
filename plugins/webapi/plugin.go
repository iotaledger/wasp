package webapi

import (
	"context"
	"errors"
	"github.com/iotaledger/wasp/packages/shutdown"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/labstack/echo"
)

// PluginName is the name of the web API plugin.
const PluginName = "WebAPI"

var (
	// Plugin is the plugin instance of the web API plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	// Server is the web API server.
	Server = echo.New()

	log *logger.Logger
)

func configure(*node.Plugin) {
	log = logger.NewLogger(PluginName)
	Server.HideBanner = true
	Server.HidePort = true
	addEndpoints()
}

func run(_ *node.Plugin) {
	log.Infof("Starting %s ...", PluginName)
	if err := daemon.BackgroundWorker("WebAPI Server", worker, shutdown.PriorityWebAPI); err != nil {
		log.Errorf("Error starting as daemon: %s", err)
	}
}

func worker(shutdownSignal <-chan struct{}) {
	defer log.Infof("Stopping %s ... done", PluginName)

	stopped := make(chan struct{})
	bindAddr := config.Node.GetString(CfgBindAddress)
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
