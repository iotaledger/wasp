package dispatcher

import (
	wasp_events "github.com/iotaledger/goshimmer/plugins/wasp/events"
	"github.com/iotaledger/goshimmer/plugins/wasp/nodeconn"
	"github.com/iotaledger/goshimmer/plugins/wasp/peering"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
)

// start qnode dispatcher daemon worker.
// It serializes all incoming 'TransactionReceived' events
func Start() {
	chMsgData := make(chan []byte)

	processNodeDataClosure := events.NewClosure(func(data []byte) {
		chMsgData <- data
	})

	processPeerMsgClosure := events.NewClosure(func(msg *wasp_events.PeerMessage) {
		if committee := committeeByColor(&msg.ScColor); committee != nil {
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
		wasp_events.Events.SCDataLoaded.Trigger()

		// goroutine to read incoming messages from the node
		go func() {
			for data := range chMsgData {
				processMsgData(data)
			}
		}()

		<-shutdownSignal

		// starting async cleanup on shutdown
		go func() {
			wasp_events.Events.NodeMessageReceived.Detach(processNodeDataClosure)
			close(chMsgData)
			log.Infof("dispatcher stopped")
		}()
	})
	if err != nil {
		log.Errorf("failed to initialize dispatcher")
		return
	}
	wasp_events.Events.NodeMessageReceived.Attach(processNodeDataClosure)
	wasp_events.Events.PeerMessageReceived.Attach(processPeerMsgClosure)

	log.Infof("dispatcher started")
}
