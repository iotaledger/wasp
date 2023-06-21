package root

import (
	"io"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// ContractRecord is a structure which contains metadata of the deployed contract instance
type ContractRecord struct {
	// The ProgramHash uniquely defines the program of the smart contract
	// It is interpreted either as one of builtin contracts (including examples)
	// or a hash (reference) to the of the blob in 'blob' contract in the 'program binary' format,
	// i.e. with at least 2 pre-defined fields:
	//  - VarFieldVType
	//  - VarFieldProgramBinary
	ProgramHash hashing.HashValue
	// Unique name of the contract on the chain. The real identity of the instance on the chain
	// is hname(name) =  isc.Hn(name)
	Name string
}

func ContractRecordFromContractInfo(itf *coreutil.ContractInfo) *ContractRecord {
	return &ContractRecord{
		ProgramHash: itf.ProgramHash,
		Name:        itf.Name,
	}
}

func ContractRecordFromBytes(data []byte) (*ContractRecord, error) {
	return rwutil.ReadFromBytes(data, new(ContractRecord))
}

func (p *ContractRecord) Bytes() []byte {
	return rwutil.WriteToBytes(p)
}

func (p *ContractRecord) Hname() isc.Hname {
	return isc.Hn(p.Name)
}

func (p *ContractRecord) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(p.ProgramHash[:])
	p.Name = rr.ReadString()
	return rr.Err
}

func (p *ContractRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(p.ProgramHash[:])
	ww.WriteString(p.Name)
	return ww.Err
}
