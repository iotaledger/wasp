// Package utils provides utility functions and interfaces for state management operations.
package utils

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type NodeRandomiser interface {
	UpdateNodeIDs([]gpa.NodeID)
	IsInitted() bool
	GetRandomOtherNodeIDs(int) []gpa.NodeID
}
