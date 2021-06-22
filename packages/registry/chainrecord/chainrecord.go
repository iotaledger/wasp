package chainrecord

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
)

// ChainRecord represents chain the node is participating in
// TODO optimize, no need for a persistent structure, simple activity tag is enough
type ChainRecord struct {
	ChainID *chainid.ChainID
	Peers   []string
	Active  bool
}

func ChainRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainRecord, error) {
	ret := &ChainRecord{}
	aliasAddr, err := ledgerstate.AliasAddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.ChainID = chainid.NewChainID(aliasAddr)

	ret.Active, err = mu.ReadBool()
	if err != nil {
		return nil, err
	}
	numPeers, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret.Peers = make([]string, numPeers)
	for i := uint16(0); i < numPeers; i++ {
		strSize, err := mu.ReadUint16()
		if err != nil {
			return nil, err
		}
		d, err := mu.ReadBytes(int(strSize))
		if err != nil {
			return nil, err
		}
		ret.Peers[i] = string(d)
	}
	return ret, nil
}

// CommitteeRecordFromBytes
func ChainRecordFromBytes(data []byte) (*ChainRecord, error) {
	return ChainRecordFromMarshalUtil(marshalutil.New(data))
}

func (rec *ChainRecord) Bytes() []byte {
	mu := marshalutil.New().WriteBytes(rec.ChainID.Bytes()).
		WriteBool(rec.Active).
		WriteUint16(uint16(len(rec.Peers)))
	for _, s := range rec.Peers {
		b := []byte(s)
		mu.WriteUint16(uint16(len(b))).
			WriteBytes(b)
	}
	return mu.Bytes()
}

func (rec *ChainRecord) String() string {
	ret := "ChainID: " + rec.ChainID.String() + "\n"
	ret += fmt.Sprintf("      Peers:  %+v\n", rec.Peers)
	ret += fmt.Sprintf("      Active: %v\n", rec.Active)
	return ret
}
