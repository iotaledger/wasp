package root

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
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
	// Description of the instance
	Description string
	// Unique name of the contract on the chain. The real identity of the instance on the chain
	// is hname(name) =  isc.Hn(name)
	Name string
}

func ContractRecordFromContractInfo(itf *coreutil.ContractInfo) *ContractRecord {
	return &ContractRecord{
		ProgramHash: itf.ProgramHash,
		Description: itf.Description,
		Name:        itf.Name,
	}
}

func ContractRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ContractRecord, error) {
	ret := &ContractRecord{}
	buf, err := mu.ReadBytes(len(ret.ProgramHash))
	if err != nil {
		return nil, err
	}
	copy(ret.ProgramHash[:], buf)

	if ret.Description, err = readString(mu); err != nil {
		return nil, err
	}
	if ret.Name, err = readString(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *ContractRecord) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(p.ProgramHash[:])
	writeString(mu, p.Description)
	writeString(mu, p.Name)
	return mu.Bytes()
}

func ContractRecordFromBytes(data []byte) (*ContractRecord, error) {
	return ContractRecordFromMarshalUtil(marshalutil.New(data))
}

func writeString(mu *marshalutil.MarshalUtil, str string) {
	mu.WriteUint16(uint16(len(str))).WriteBytes([]byte(str))
}

func readString(mu *marshalutil.MarshalUtil) (string, error) {
	sz, err := mu.ReadUint16()
	if err != nil {
		return "", err
	}
	ret, err := mu.ReadBytes(int(sz))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func (p *ContractRecord) Hname() isc.Hname {
	return isc.Hn(p.Name)
}
