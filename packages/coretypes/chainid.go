package coretypes

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/mr-tron/base58"
)

const ChainIDLength = address.Length

type ChainID address.Address // in the future type ChainIDs balance.Color

var NilChainID = ChainID{}

func NewChainIDFromBase58(b58 string) (ret ChainID, err error) {
	var b []byte
	b, err = base58.Decode(b58)
	if err != nil {
		return
	}
	if len(b) != ChainIDLength {
		err = ErrWrongDataLength
		return
	}
	copy(ret[:], b)
	return
}

func NewChainIDFromBytes(data []byte) (ret ChainID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// RandomChainID creates a random chain ID.
func RandomChainID() ChainID {
	return (ChainID)(address.Random())
}

func (chid ChainID) String() string {
	return (address.Address)(chid).String()
}

func (chid *ChainID) Write(w io.Writer) error {
	_, err := w.Write(chid[:])
	return err
}

func (chid *ChainID) Read(r io.Reader) error {
	n, err := r.Read(chid[:])
	if err != nil {
		return err
	}
	if n != ChainIDLength {
		return ErrWrongDataLength
	}
	return nil
}
