package root

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type ContractRecord struct {
	VMType         string
	DeploymentHash hashing.HashValue // hash(VMType, program binary)
	Description    string
	NodeFee        int64 // minimum node fee
}

// serde
func (p *ContractRecord) Write(w io.Writer) error {
	if err := util.WriteString16(w, p.VMType); err != nil {
		return err
	}
	if _, err := w.Write(p.DeploymentHash[:]); err != nil {
		return err
	}
	if err := util.WriteString16(w, p.Description); err != nil {
		return err
	}
	if err := util.WriteInt64(w, p.NodeFee); err != nil {
		return err
	}
	return nil
}

func (p *ContractRecord) Read(r io.Reader) error {
	var err error
	if p.VMType, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadHashValue(r, &p.DeploymentHash); err != nil {
		return err
	}
	if p.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &p.NodeFee); err != nil {
		return err
	}
	return nil
}

func EncodeContractRecord(p *ContractRecord) []byte {
	return util.MustBytes(p)
}

func DecodeContractRecord(data []byte) (*ContractRecord, error) {
	ret := new(ContractRecord)
	err := ret.Read(bytes.NewReader(data))
	return ret, err
}

func GetRootContractRecord() *ContractRecord {
	return &ContractRecord{
		VMType:         "builtin",
		DeploymentHash: *ProgramHash,
		Description:    "root contract",
	}
}
