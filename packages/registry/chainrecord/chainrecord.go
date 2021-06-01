package chainrecord

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
)

// ChainRecord represents chain the node is participating in
// TODO optimize, no need for a persistent structure, simple activity tag is enough
type ChainRecord struct {
	ChainAddr *ledgerstate.AliasAddress
	Active    bool
}

func NewChainRecord(chainID *ledgerstate.AliasAddress, active bool) *ChainRecord {
	return &ChainRecord{
		ChainAddr: ledgerstate.NewAliasAddress(chainID.Bytes()),
		Active:    active,
	}
}

func ChainRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainRecord, error) {
	ret := &ChainRecord{}
	aliasAddr, err := ledgerstate.AliasAddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.ChainAddr = aliasAddr

	ret.Active, err = mu.ReadBool()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// CommitteeRecordFromBytes
func ChainRecordFromBytes(data []byte) (*ChainRecord, error) {
	return ChainRecordFromMarshalUtil(marshalutil.New(data))
}

func (rec *ChainRecord) Bytes() []byte {
	return marshalutil.New().
		WriteBytes(rec.ChainAddr.Bytes()).
		WriteBool(rec.Active).
		Bytes()
}

func (rec *ChainRecord) String() string {
	ret := "ChainID: " + rec.ChainAddr.String() + "\n"
	ret += fmt.Sprintf("      Active: %v\n", rec.Active)
	return ret
}
