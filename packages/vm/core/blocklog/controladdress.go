package blocklog

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// region ControlAddresses ///////////////////////////////////////////////

type ControlAddresses struct {
	StateAddress     iotago.Address
	GoverningAddress iotago.Address
	SinceBlockIndex  uint32
}

func ControlAddressesFromBytes(data []byte) (*ControlAddresses, error) {
	return rwutil.ReaderFromBytes(data, new(ControlAddresses))
}

func (ca *ControlAddresses) Bytes() []byte {
	return rwutil.WriterToBytes(ca)
}

func (ca *ControlAddresses) String() string {
	var ret string
	if ca.StateAddress.Equal(ca.GoverningAddress) {
		ret = fmt.Sprintf("ControlAddresses(%s), block: %d", ca.StateAddress, ca.SinceBlockIndex)
	} else {
		ret = fmt.Sprintf("ControlAddresses(%s, %s), block: %d",
			ca.StateAddress, ca.GoverningAddress, ca.SinceBlockIndex)
	}
	return ret
}

func (ca *ControlAddresses) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	ca.StateAddress = isc.AddressFromReader(rr)
	ca.GoverningAddress = isc.AddressFromReader(rr)
	ca.SinceBlockIndex = rr.ReadUint32()
	return rr.Err
}

func (ca *ControlAddresses) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	isc.AddressToWriter(ww, ca.StateAddress)
	isc.AddressToWriter(ww, ca.GoverningAddress)
	ww.WriteUint32(ca.SinceBlockIndex)
	return ww.Err
}

// endregion /////////////////////////////////////////////////////////////
