// +build windows

package cluster

import (
	"golang.org/x/xerrors"
)

func (clu *Cluster) FreezeNode(nodeIndex int) error {
	return xerrors.Errorf("Freezing is not supported on Windows")
}

func (clu *Cluster) UnfreezeNode(nodeIndex int) error {
	return xerrors.Errorf("Freezing is not supported on Windows")
}
