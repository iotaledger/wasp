package coretypes

import (
	"bytes"
	"fmt"
	"github.com/mr-tron/base58"
	"io"
)

const ContractIDLength = ChainIDLength + 2

type ContractID [ContractIDLength]byte

func NewContractID(chid ChainID, index Uint16) (ret ContractID) {
	copy(ret[:ChainIDLength], chid[:])
	copy(ret[ChainIDLength:], index.Bytes())
	return
}

func NewContractIDFromBytes(data []byte) (ret ContractID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

func NewContractIDFromBase58(base58string string) (ret ContractID, err error) {
	var data []byte
	if data, err = base58.Decode(base58string); err != nil {
		return
	}
	return NewContractIDFromBytes(data)
}

func (scid ContractID) ChainID() (ret ChainID) {
	copy(ret[:ChainIDLength], scid[:ChainIDLength])
	return
}

func (scid ContractID) Index() (ret Uint16) {
	ret, _ = NewUint16From2Bytes(scid[ChainIDLength:])
	return
}

func (scid ContractID) Base58() string {
	return base58.Encode(scid[:])
}

const (
	long_format  = "%s::%d"
	short_format = "%s..::%d"
)

func (scid ContractID) String() string {
	return fmt.Sprintf(long_format, scid.ChainID().String(), scid.Index())
}

func (scid ContractID) Short() string {
	return fmt.Sprintf(short_format, scid.ChainID().String()[:8], scid.Index())
}

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

func (scid *ContractID) Write(w io.Writer) error {
	_, err := w.Write(scid[:])
	return err
}
