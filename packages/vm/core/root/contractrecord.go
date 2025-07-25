package root

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
)

// ContractRecord is a structure which contains metadata of the deployed contract instance
type ContractRecord struct {
	// Unique name of the contract on the chain. The real identity of the instance on the chain
	// is hname(name) =  isc.Hn(name)
	Name string
}

func ContractRecordFromContractInfo(itf *coreutil.ContractInfo) *ContractRecord {
	return &ContractRecord{
		Name: itf.Name,
	}
}

func ContractRecordFromBytes(data []byte) (*ContractRecord, error) {
	return bcs.Unmarshal[*ContractRecord](data)
}

func (p *ContractRecord) Bytes() []byte {
	return bcs.MustMarshal(p)
}

func (p *ContractRecord) Hname() isc.Hname {
	return isc.Hn(p.Name)
}
