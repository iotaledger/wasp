package metrics

import "github.com/iotaledger/wasp/v2/packages/isc"

type ChainStateLag map[isc.ChainID]*chainSyncLagEntry

type chainSyncLagEntry struct {
	wantStateIndex uint32
	haveStateIndex uint32
}

func (ce chainSyncLagEntry) Lag() uint32 {
	return ce.wantStateIndex - ce.haveStateIndex
}

func (c ChainStateLag) Want(chainID isc.ChainID, stateIndex uint32) {
	if e, ok := c[chainID]; ok {
		e.wantStateIndex = stateIndex
		return
	}
	c[chainID] = &chainSyncLagEntry{wantStateIndex: stateIndex}
}

func (c ChainStateLag) Have(chainID isc.ChainID, stateIndex uint32) {
	if e, ok := c[chainID]; ok {
		e.haveStateIndex = stateIndex
		return
	}
	c[chainID] = &chainSyncLagEntry{haveStateIndex: stateIndex}
}

func (c ChainStateLag) ChainLag(chainID isc.ChainID) uint32 {
	if e, ok := c[chainID]; ok {
		return e.Lag()
	}
	return 0
}

func (c ChainStateLag) MaxLag() uint32 {
	maxLag := uint32(0)
	for _, e := range c {
		lag := e.Lag()
		if maxLag < lag {
			maxLag = lag
		}
	}
	return maxLag
}
