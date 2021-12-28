package nodeconnimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/txstream"
	txstream_client "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type nodeConnImplementation struct {
	client                 *txstream_client.Client
	transactionHandlers    map[ledgerstate.AliasAddress]chain.NodeConnectionHandleTransactionFun
	iStateHandlers         map[ledgerstate.AliasAddress]chain.NodeConnectionHandleInclusionStateFun
	outputHandlers         map[ledgerstate.AliasAddress]chain.NodeConnectionHandleOutputFun
	unspentAOutputHandlers map[ledgerstate.AliasAddress]chain.NodeConnectionHandleUnspentAliasOutputFun
	transactionClosure     *events.Closure
	iStateClosure          *events.Closure
	outputClosure          *events.Closure
	unspentAOutputClosure  *events.Closure
	metrics                nodeconnmetrics.NodeConnectionMetrics
	log                    *logger.Logger // general chains logger
}

var _ chain.NodeConnection = &nodeConnImplementation{}

func NewNodeConnection(nodeConnClient *txstream_client.Client, metrics nodeconnmetrics.NodeConnectionMetrics, log *logger.Logger) chain.NodeConnection {
	ret := &nodeConnImplementation{
		client:                 nodeConnClient,
		transactionHandlers:    make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleTransactionFun),
		iStateHandlers:         make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleInclusionStateFun),
		outputHandlers:         make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleOutputFun),
		unspentAOutputHandlers: make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleUnspentAliasOutputFun),
		metrics:                metrics,
		log:                    log,
	}

	ret.transactionClosure = events.NewClosure(ret.handleTransactionReceived)
	ret.client.Events.TransactionReceived.Attach(ret.transactionClosure)

	ret.iStateClosure = events.NewClosure(ret.handleInclusionStateReceived)
	ret.client.Events.InclusionStateReceived.Attach(ret.iStateClosure)

	ret.outputClosure = events.NewClosure(ret.handleOutputReceived)
	ret.client.Events.OutputReceived.Attach(ret.outputClosure)

	ret.unspentAOutputClosure = events.NewClosure(ret.handleUnspentAliasOutputReceived)
	ret.client.Events.UnspentAliasOutputReceived.Attach(ret.unspentAOutputClosure)

	return ret
}

func (n *nodeConnImplementation) handleTransactionReceived(msg *txstream.MsgTransaction) {
	n.log.Debugf("NodeConnnection::TransactionReceived...")
	defer n.log.Debugf("NodeConnnection::TransactionReceived... Done")
	n.metrics.GetInTransaction().CountLastMessage(msg)
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		n.log.Warnf("NodeConnnection::TransactionReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
		return
	}
	handler, ok := n.transactionHandlers[*aliasAddr]
	if !ok {
		n.log.Warnf("NodeConnnection::TransactionReceived: no handler for address %v", aliasAddr.String())
		return
	}
	handler(msg.Tx)
}

func (n *nodeConnImplementation) handleInclusionStateReceived(msg *txstream.MsgTxInclusionState) {
	n.log.Debugf("NodeConnnection::InclusionStateReceived...")
	defer n.log.Debugf("NodeConnnection::InclusionStateReceived... Done")
	n.metrics.GetInInclusionState().CountLastMessage(msg)
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		n.log.Warnf("NodeConnnection::InclusionStateReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
		return
	}
	handler, ok := n.iStateHandlers[*aliasAddr]
	if !ok {
		n.log.Warnf("NodeConnnection::InclusionStateReceived: no handler for address %v", aliasAddr.String())
		return
	}
	handler(msg.TxID, msg.State)
}

func (n *nodeConnImplementation) handleOutputReceived(msg *txstream.MsgOutput) {
	n.log.Debugf("NodeConnnection::OutputReceived...")
	defer n.log.Debugf("NodeConnnection::OutputReceived... Done")
	n.metrics.GetInOutput().CountLastMessage(msg)
	aliasAddr, ok := msg.Address.(*ledgerstate.AliasAddress)
	if !ok {
		n.log.Warnf("NodeConnnection::OutputReceived: cannot dispatch transaction message to non-alias address %v", msg.Address.String())
		return
	}
	handler, ok := n.outputHandlers[*aliasAddr]
	if !ok {
		n.log.Warnf("NodeConnnection::OutputReceived: no handler for address %v", aliasAddr.String())
		return
	}
	handler(msg.Output)
}

func (n *nodeConnImplementation) handleUnspentAliasOutputReceived(msg *txstream.MsgUnspentAliasOutput) {
	n.log.Debugf("NodeConnnection::UnspentAliasOutputReceived...")
	defer n.log.Debugf("NodeConnnection::UnspentAliasOutputReceived... Done")
	n.metrics.GetInUnspentAliasOutput().CountLastMessage(msg)
	handler, ok := n.unspentAOutputHandlers[*msg.AliasAddress]
	if !ok {
		n.log.Warnf("NodeConnnection::UnspentAliasOutputReceived: no handler for address %v", msg.AliasAddress.String())
		return
	}
	handler(msg.AliasOutput, msg.Timestamp)
}

// NOTE: request to client methods are logged through each chain logger in chainNodeConnImplementation

func (n *nodeConnImplementation) PullState(addr *ledgerstate.AliasAddress) {
	n.metrics.GetOutPullState().CountLastMessage(addr)
	n.client.RequestUnspentAliasOutput(addr)
}

func (n *nodeConnImplementation) PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID) {
	n.metrics.GetOutPullTransactionInclusionState().CountLastMessage(struct {
		Address       ledgerstate.Address
		TransactionID ledgerstate.TransactionID
	}{
		Address:       addr,
		TransactionID: txid,
	})
	n.client.RequestTxInclusionState(addr, txid)
}

func (n *nodeConnImplementation) PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
	n.metrics.GetOutPullConfirmedOutput().CountLastMessage(struct {
		Address  ledgerstate.Address
		OutputID ledgerstate.OutputID
	}{
		Address:  addr,
		OutputID: outputID,
	})
	n.client.RequestConfirmedOutput(addr, outputID)
}

func (n *nodeConnImplementation) PostTransaction(tx *ledgerstate.Transaction) {
	n.metrics.GetOutPostTransaction().CountLastMessage(tx)
	n.client.PostTransaction(tx)
}

func (n *nodeConnImplementation) AttachToTransactionReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleTransactionFun) {
	n.log.Debugf("NodeConnnection::AttachToTransactionReceived to %v", addr.String())
	_, ok := n.transactionHandlers[*addr]
	if ok {
		n.log.Panicf("NodeConnnection::AttachToTransactionReceived to %v failed: handler already registered")
	}
	n.transactionHandlers[*addr] = handler
}

func (n *nodeConnImplementation) AttachToInclusionStateReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleInclusionStateFun) {
	n.log.Debugf("NodeConnnection::AttachToInclusionStateReceived to %v", addr.String())
	_, ok := n.iStateHandlers[*addr]
	if ok {
		n.log.Panicf("NodeConnnection::AttachToInclusionStateReceived to %v failed: handler already registered")
	}
	n.iStateHandlers[*addr] = handler
}

func (n *nodeConnImplementation) AttachToOutputReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleOutputFun) {
	n.log.Debugf("NodeConnnection::AttachToOutputReceived to %v", addr.String())
	_, ok := n.outputHandlers[*addr]
	if ok {
		n.log.Panicf("NodeConnnection::AttachToOutputReceived to %v failed: handler already registered")
	}
	n.outputHandlers[*addr] = handler
}

func (n *nodeConnImplementation) AttachToUnspentAliasOutputReceived(addr *ledgerstate.AliasAddress, handler chain.NodeConnectionHandleUnspentAliasOutputFun) {
	n.log.Debugf("NodeConnnection::AttachToUnspentAliasOutputReceived to %v", addr.String())
	_, ok := n.unspentAOutputHandlers[*addr]
	if ok {
		n.log.Panicf("NodeConnnection::AttachToUnspentAliasOutputReceived to %v failed: handler already registered")
	}
	n.unspentAOutputHandlers[*addr] = handler
}

func (n *nodeConnImplementation) DetachFromTransactionReceived(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnnection::DetachFromTransactionReceived to %v", addr.String())
	delete(n.transactionHandlers, *addr)
}

func (n *nodeConnImplementation) DetachFromInclusionStateReceived(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnnection::DetachFromInclusionStateReceived to %v", addr.String())
	delete(n.iStateHandlers, *addr)
}

func (n *nodeConnImplementation) DetachFromOutputReceived(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnnection::DetachFromOutputReceived to %v", addr.String())
	delete(n.outputHandlers, *addr)
}

func (n *nodeConnImplementation) DetachFromUnspentAliasOutputReceived(addr *ledgerstate.AliasAddress) {
	n.log.Debugf("NodeConnnection::DetachFromUnspentAliasOutputReceived to %v", addr.String())
	delete(n.unspentAOutputHandlers, *addr)
}

func (n *nodeConnImplementation) Subscribe(addr ledgerstate.Address) {
	n.log.Debugf("NodeConnnection::Subscribing to %v...", addr.String())
	defer n.log.Debugf("NodeConnnection::Subscribing done")
	n.client.Subscribe(addr)
	n.metrics.SetSubscribed(addr)
}

func (n *nodeConnImplementation) Unsubscribe(addr ledgerstate.Address) {
	n.log.Debugf("NodeConnnection::Unsubscribing from %v...", addr.String())
	defer n.log.Debugf("NodeConnnection::Unsubscribing done")
	n.metrics.SetUnsubscribed(addr)
	n.client.Unsubscribe(addr)
}

func (n *nodeConnImplementation) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return n.metrics
}

func (n *nodeConnImplementation) Close() {
	n.log.Debugf("NodeConnnection::Close")
	n.client.Events.TransactionReceived.Detach(n.transactionClosure)
	n.transactionHandlers = make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleTransactionFun)

	n.client.Events.InclusionStateReceived.Detach(n.iStateClosure)
	n.iStateHandlers = make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleInclusionStateFun)

	n.client.Events.OutputReceived.Detach(n.outputClosure)
	n.outputHandlers = make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleOutputFun)

	n.client.Events.UnspentAliasOutputReceived.Detach(n.unspentAOutputClosure)
	n.unspentAOutputHandlers = make(map[ledgerstate.AliasAddress]chain.NodeConnectionHandleUnspentAliasOutputFun)
}
