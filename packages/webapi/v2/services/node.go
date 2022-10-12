package services

import (
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type NodeService struct {
	logger *logger.Logger

	shutdownHandler *shutdown.ShutdownHandler
}

func NewNodeService(log *logger.Logger, shutdownHandler *shutdown.ShutdownHandler) interfaces.NodeService {
	return &NodeService{
		logger: log,

		shutdownHandler: shutdownHandler,
	}
}

func (n *NodeService) ShutdownNode() {
	n.shutdownHandler.SelfShutdown("wasp was shutdown via API", false)
}
