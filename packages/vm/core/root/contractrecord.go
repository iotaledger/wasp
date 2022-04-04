package root

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
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
	// is hname(name) =  iscp.Hn(name)
	Name string
	// The agentID of the entity which deployed the instance. It can be interpreted as
	// an priviledged user of the instance, however it is up to the smart contract.
	Creator *iscp.AgentID
}

func ContractRecordFromContractInfo(itf *coreutil.ContractInfo, creator *iscp.AgentID) *ContractRecord {
	if creator == nil {
		panic("ContractRecordFromContractInfo: creator can't be nil")
	}
	return &ContractRecord{
		ProgramHash: itf.ProgramHash,
		Description: itf.Description,
		Name:        itf.Name,
		Creator:     creator,
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
	creatorNotNil, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if creatorNotNil {
		if ret.Creator, err = iscp.AgentIDFromMarshalUtil(mu); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (p *ContractRecord) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBytes(p.ProgramHash[:])
	writeString(mu, p.Description)
	writeString(mu, p.Name)
	mu.WriteBool(p.Creator != nil)
	if p.Creator != nil {
		mu.Write(p.Creator)
	}
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

func (p *ContractRecord) HasCreator() bool {
	return p.Creator != nil
}
