package chain

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"math/rand"
)

type PeerGroup struct {
	PeerGroupProvider
}

func NewPeerGroup(peers PeerGroupProvider) *PeerGroup {
	return &PeerGroup{peers}
}

func (pgT *PeerGroup) SendMsgToRandomNodes(number int, msgType byte, msgData []byte) error {
	pgT.Lock()
	defer pgT.Unlock()

	var rndBytes [32]byte
	rand.Read(rndBytes[:])
	permutation := util.NewPermutation16(pgT.NumPeers(), rndBytes[:]).GetArray()
	sent := 0
	for _, index := range permutation {
		err := pgT.SendMsg(index, msgType, msgData)
		if err == nil {
			sent++
			if sent == number {
				return nil
			}
		}
	}
	return fmt.Errorf("Sent to %v nodes only", sent)
}
