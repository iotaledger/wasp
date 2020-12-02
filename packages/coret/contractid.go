package coret

import (
	"bytes"
	"fmt"
	"github.com/mr-tron/base58"
	"io"
)

const ContractIDLength = ChainIDLength + HnameLength

// ContractID global identifier of the smart contract
type ContractID [ContractIDLength]byte

// NewContractID a constructor
func NewContractID(chid ChainID, contractHn Hname) (ret ContractID) {
	copy(ret[:ChainIDLength], chid[:])
	copy(ret[ChainIDLength:], contractHn.Bytes())
	return
}

// NewContractIDFromBytes a constructor
func NewContractIDFromBytes(data []byte) (ret ContractID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewContractIDFromBase58 a constructor, unmarshals base58 string
func NewContractIDFromBase58(base58string string) (ret ContractID, err error) {
	var data []byte
	if data, err = base58.Decode(base58string); err != nil {
		return
	}
	return NewContractIDFromBytes(data)
}

// ChainID ID of the native chain of the contract
func (scid ContractID) ChainID() (ret ChainID) {
	copy(ret[:ChainIDLength], scid[:ChainIDLength])
	return
}

// Hname hashed name of the contract, local ID on the chain
func (scid ContractID) Hname() Hname {
	ret, _ := NewHnameFromBytes(scid[ChainIDLength:])
	return ret
}

// Base58 base58 representation of the binary date
func (scid ContractID) Base58() string {
	return base58.Encode(scid[:])
}

const (
	long_format  = "%s::%s"
	short_format = "%s..::%s"
)

// String human readable representation
func (scid ContractID) String() string {
	return fmt.Sprintf(long_format, scid.ChainID().String(), scid.Hname().String())
}

// Short human readable short representation
func (scid ContractID) Short() string {
	return fmt.Sprintf(short_format, scid.ChainID().String()[:8], scid.Hname().String())
}

// Read unmarshal
func (scid *ContractID) Read(r io.Reader) error {
	n, err := r.Read(scid[:])
	if err != nil {
		return err
	}
	if n != ContractIDLength {
		return ErrWrongDataLength
	}
	return nil
}

// Write marshal
func (scid *ContractID) Write(w io.Writer) error {
	_, err := w.Write(scid[:])
	return err
}
