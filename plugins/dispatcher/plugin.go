// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package dispatcher router goshimmer node messages to the corresponding
// components in the wasp node.
package dispatcher

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	_ "github.com/iotaledger/wasp/packages/chain/chainimpl" // activate init
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/nodeconn"
)

const PluginName = "Dispatcher"

var (
	log *logger.Logger
)

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	state.InitLogger()
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {

		chNodeMsg := make(chan interface{}, 100)

		processNodeMsgClosure := events.NewClosure(func(msg interface{}) {
			chNodeMsg <- msg
		})

		err := daemon.BackgroundWorker("wasp dispatcher", func(shutdownSignal <-chan struct{}) {
			// goroutine to read incoming messages from the node
			go func() {
				for msg := range chNodeMsg {
					processNodeMsg(msg)
				}
			}()

			<-shutdownSignal

			log.Infof("Stopping %s..", PluginName)
			go func() {
				nodeconn.NodeConn.Events.MessageReceived.Detach(processNodeMsgClosure)

				close(chNodeMsg)
				log.Infof("Stopping %s.. Done", PluginName)
			}()
		})
		if err != nil {
			log.Errorf("failed to initialize %v", PluginName)
			return
		}

		// event attachments
		// receiving events from NodeConn --> producing dispatcher events
		nodeconn.NodeConn.Events.MessageReceived.Attach(processNodeMsgClosure)

		log.Infof("dispatcher started")

	}, parameters.PriorityDispatcher)

	if err != nil {
		log.Errorf("failed to start worker for %s: %v", PluginName, err)
	}
}

func processNodeMsg(msg interface{}) {
	switch msgt := msg.(type) {
	case *waspconn.WaspFromNodeTransactionMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		chain := chains.GetChain(chainID)
		if chain == nil {
			return
		}
		log.Debugw("dispatch transaction",
			"txid", msgt.Tx.ID().String(),
			"chainid", chainID.String(),
		)
		chain.ReceiveMessage(msgt)

	case *waspconn.WaspFromNodeTxInclusionStateMsg:
		chainID := coretypes.NewChainID(msgt.ChainAddress)
		ch := chains.GetChain(chainID)
		if ch == nil {
			return
		}
		ch.ReceiveMessage(msgt)
	}
	log.Errorf("wrong message type")
}
