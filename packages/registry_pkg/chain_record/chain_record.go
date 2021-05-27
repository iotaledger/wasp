package chain_record

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
)

// ChainRecord represents chain the node is participating in
// TODO optimize, no need for a persistent structure, simple activity tag is enough
type ChainRecord struct {
	ChainIdAliasAddress *ledgerstate.AliasAddress
	Active              bool
	DedicatedDbInstance bool // whether the chain data is stored as a separate db instance/file
}

func NewChainRecord(chainID *ledgerstate.AliasAddress, active ...bool) *ChainRecord {
	act := false
	if len(active) > 0 {
		act = active[0]
	}
	return &ChainRecord{
		ChainIdAliasAddress: ledgerstate.NewAliasAddress(chainID.Bytes()),
		Active:              act,
	}
}

func ChainRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainRecord, error) {
	ret := &ChainRecord{}
	aliasAddr, err := ledgerstate.AliasAddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.ChainIdAliasAddress = aliasAddr

	ret.Active, err = mu.ReadBool()
	if err != nil {
		return nil, err
	}

	ret.DedicatedDbInstance, err = mu.ReadBool()
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
		WriteBytes(rec.ChainIdAliasAddress.Bytes()).
		WriteBool(rec.Active).
		WriteBool(rec.DedicatedDbInstance).
		Bytes()
}

func (rec *ChainRecord) String() string {
	ret := "ChainID: " + rec.ChainIdAliasAddress.String() + "\n"
	ret += fmt.Sprintf("      Active: %v\n", rec.Active)
	return ret
}
