// package for maintaining connection with the main node
// on the node WaspConn plugin is handling yhe connection
package nodeconn

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/chopper"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/payload"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/plugins/peering"
)

const (
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
	retryAfter   = 8 * time.Second
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

// dials outbound address and established connection
func nodeConnect() {
	addr := parameters.GetString(parameters.NodeAddress)
	log.Infof("connecting with node at %s", addr)

	var conn net.Conn
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return fmt.Errorf("can't connect with the node: %v", err)
		}
		return nil
	}); err != nil {
		log.Warn(err)

		retryNodeConnect()
		return
	}

	bconnMutex.Lock()
	bconn = buffconn.NewBufferedConnection(conn, payload.MaxMessageSize)
	bconnMutex.Unlock()

	log.Debugf("established connection with node at %s", addr)

	dataReceivedClosure := events.NewClosure(func(data []byte) {
		msgDataToEvent(data)
	})

	bconn.Events.ReceiveMessage.Attach(dataReceivedClosure)
	bconn.Events.Close.Attach(events.NewClosure(func() {
		log.Errorf("lost connection with %s", addr)
		go func() {
			bconnMutex.Lock()
			bconnSave := bconn
			bconn = nil
			bconnMutex.Unlock()
			bconnSave.Events.ReceiveMessage.Detach(dataReceivedClosure)
		}()
	}))

	if err := SendWaspIdToNode(); err == nil {
		log.Debugf("sent own wasp id to node: %s", peering.DefaultNetworkProvider().Self().Location())
	} else {
		log.Errorf("failed to send wasp id to node: %v", err)
	}

	// read loop
	if err := bconn.Read(); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Warnw("Permanent error", "err", err)
		}
	}
	// try to reconnect after some time
	log.Debugf("disconnected from node. Will try to reconnect after %v", retryAfter)

	retryNodeConnect()
}

func IsConnected() bool {
	bconnMutex.Lock()
	defer bconnMutex.Unlock()
	return bconn != nil
}

func retryNodeConnect() {
	log.Infof("will retry connecting to the node after %v", retryAfter)
	time.Sleep(retryAfter)
	go nodeConnect()
}

func SendDataToNode(data []byte) error {
	choppedData, chopped := chopper.ChopData(data, payload.MaxMessageSize-waspconn.ChunkMessageHeaderSize)

	bconnMutex.Lock()
	defer bconnMutex.Unlock()

	if bconn == nil {
		return fmt.Errorf("SendDataToNode: not connected to node")
	}
	var err error
	if !chopped {
		_, err = bconn.Write(data)
	} else {
		for _, piece := range choppedData {
			var d []byte
			d, err = waspconn.EncodeMsg(&waspconn.WaspMsgChunk{
				Data: piece,
			})
			if err != nil {
				break
			}
			_, err = bconn.Write(d)
			if err != nil {
				break
			}
		}
	}
	return err
}
