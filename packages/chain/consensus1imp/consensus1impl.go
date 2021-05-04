package consensus1imp

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
)

type consensusImpl struct {
	chain     chain.ChainCore
	committee chain.Committee
	mempool   chain.Mempool
	nodeConn  chain.NodeConnection
}

var _ chain.Consensus1 = &consensusImpl{}

func New(chainCore chain.ChainCore, mempool chain.Mempool, committee chain.Committee, nodeConn chain.NodeConnection, log *logger.Logger) *consensusImpl {
	return &consensusImpl{
		chain:     chainCore,
		committee: committee,
		mempool:   mempool,
		nodeConn:  nodeConn,
	}
}
