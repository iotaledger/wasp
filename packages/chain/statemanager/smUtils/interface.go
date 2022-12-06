package smUtils

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type NodeRandomiser interface {
	UpdateNodeIDs([]gpa.NodeID)
	IsInitted() bool
	GetRandomOtherNodeIDs(int) []gpa.NodeID
}
