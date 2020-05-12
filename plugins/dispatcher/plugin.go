package dispatcher

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
)

// PluginName is the name of the database plugin.
const PluginName = "Dispatcher"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {

		chMsgData := make(chan []byte, 100)

		processNodeDataClosure := events.NewClosure(func(data []byte) {
			chMsgData <- data
		})

		processPeerMsgClosure := events.NewClosure(func(msg *peering.PeerMessage) {
			if committee := committeeByAddress(&msg.Address); committee != nil {
				committee.ReceiveMessage(msg)
			}
		})

		err := daemon.BackgroundWorker("wasp dispatcher", func(shutdownSignal <-chan struct{}) {
			// load all sc data records from registry
			addrs, err := loadAllSContracts(peering.OwnPortAddr())
			if err != nil {
				log.Error("failed to load SC data from registry: %v", err)
				return
			}
			log.Debugf("loaded %d SC data record(s) from registry", len(addrs))

			// let the node know addresses of interest
			nodeconn.SetSubscriptions(addrs)

			// trigger event to notify that SC data is initialized
			EventSCDataLoaded.Trigger()

			// goroutine to read incoming messages from the node
			go func() {
				for data := range chMsgData {
					processMsgData(data)
				}
			}()

			<-shutdownSignal

			log.Info("Stopping dispatcher..")
			go func() {
				nodeconn.EventNodeMessageReceived.Detach(processNodeDataClosure)
				close(chMsgData)
				log.Info("Stopping dispatcher.. Done")
			}()
		})
		if err != nil {
			log.Errorf("failed to initialize dispatcher")
			return
		}

		nodeconn.EventNodeMessageReceived.Attach(processNodeDataClosure)
		peering.EventPeerMessageReceived.Attach(processPeerMsgClosure)

		log.Infof("dispatcher started")

	}, shutdown.PriorityDispatcher)

	if err != nil {
		log.Errorf("failed to start worker: %v", err)
	}
}
