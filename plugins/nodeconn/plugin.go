package nodeconn

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/plugins/peering"
	"net"
	"time"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "NodeConn"

const dialTimeout = 1 * time.Second

var (
	log *logger.Logger

	NodeConn *nodeconn.NodeConn
	Ready    = ready.New()
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		addr := parameters.GetString(parameters.NodeAddress)
		dial := nodeconn.DialFunc(func() (string, net.Conn, error) {
			log.Infof("connecting with node at %s", addr)
			conn, err := net.DialTimeout("tcp", addr, dialTimeout)
			return addr, conn, err
		})

		NodeConn := nodeconn.New(peering.DefaultNetworkProvider().Self().NetID(), log, dial)
		defer NodeConn.Close()

		<-shutdownSignal

		log.Info("Stopping node connection..")
	}, parameters.PriorityNodeConnection)
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
	Ready.SetReady()
}
