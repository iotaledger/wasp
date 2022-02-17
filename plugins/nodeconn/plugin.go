package nodeconn

import (
	"context"
	"net"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/txstream"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/plugins/peering"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "NodeConn"

const dialTimeout = 1 * time.Second

var (
	log *logger.Logger

	nodeConn    *txstream.Client
	initialized = ready.New("NodeConn")
)

// Init initializes the plugin
func Init() *node.Plugin {
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

func NodeConnection() *txstream.Client {
	initialized.MustWait(5 * time.Second)
	return nodeConn
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(ctx context.Context) {
		addr := parameters.GetString(parameters.NodeAddress)
		dial := txstream.DialFunc(func() (string, net.Conn, error) {
			log.Infof("connecting with node at %s", addr)
			conn, err := net.DialTimeout("tcp", addr, dialTimeout)
			return addr, conn, err
		})

		nodeConn = txstream.New(peering.DefaultNetworkProvider().Self().NetID(), log, dial)
		initialized.SetReady()
		defer nodeConn.Close()

		<-ctx.Done()

		log.Info("Stopping node connection..")
	}, parameters.PriorityNodeConnection)
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}
