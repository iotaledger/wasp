// +build linux 
// +build darwin

package cluster

import (
	"syscall"

	"golang.org/x/xerrors"
)

func (clu *Cluster) FreezeNode(nodeIndex int) error {
	if nodeIndex >= len(clu.waspCmds) {
		return xerrors.Errorf("[cluster] Wasp node with index %d not found, active processes: %v", nodeIndex, len(clu.waspCmds))
	}

	process := clu.waspCmds[nodeIndex]

	err := process.Process.Signal(syscall.SIGSTOP)

	return err
}

func (clu *Cluster) UnfreezeNode(nodeIndex int) error {
	if nodeIndex >= len(clu.waspCmds) {
		return xerrors.Errorf("[cluster] Wasp node with index %d not found", nodeIndex)
	}

	process := clu.waspCmds[nodeIndex]

	err := process.Process.Signal(syscall.SIGCONT)

	return err
}
