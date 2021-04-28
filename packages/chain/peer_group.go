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
	fmt.Printf("XXX PERMUTATION %v\n", permutation)
	sent := 0
	for _, index := range permutation {
		fmt.Printf("XXX SENDING to %v\n", index)
		err := pgT.SendMsg(index, msgType, msgData)
		fmt.Printf("XXX SENT to %v RESULT %v\n", index, err)
		if err == nil {
			sent++
			if sent == number {
				return nil
			}
		}
	}
	fmt.Printf("XXX SENDING too few %v\n", sent)
	return fmt.Errorf("Sent to %v nodes only", sent)
}
