package blocklog

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

// region ControlAddresses ///////////////////////////////////////////////

type ControlAddresses struct {
	StateAddress     iotago.Address
	GoverningAddress iotago.Address
	SinceBlockIndex  uint32
}

func ControlAddressesFromBytes(data []byte) (*ControlAddresses, error) {
	return ControlAddressesFromMarshalUtil(marshalutil.New(data))
}

func ControlAddressesFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ControlAddresses, error) {
	ret := &ControlAddresses{}
	var err error

	if ret.StateAddress, err = isc.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.GoverningAddress, err = isc.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.SinceBlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (ca *ControlAddresses) Bytes() []byte {
	mu := marshalutil.New()

	mu.WriteBytes(isc.BytesFromAddress(ca.StateAddress)).
		WriteBytes(isc.BytesFromAddress(ca.GoverningAddress)).
		WriteUint32(ca.SinceBlockIndex)
	return mu.Bytes()
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

// endregion /////////////////////////////////////////////////////////////
