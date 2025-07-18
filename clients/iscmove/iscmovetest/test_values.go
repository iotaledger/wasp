package iscmovetest

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/testutil/testval"
	"github.com/samber/lo"
)

var TestAssetsBag = iscmove.AssetsBag{
	ID:   *iotatest.TestAddress,
	Size: 0,
}

var TestAssetBagReferent = iscmove.Referent[iscmove.AssetsBag]{
	ID:    *iotatest.TestAddress,
	Value: lo.ToPtr(TestAssetsBag),
}

var TestAnchor = RandomAnchor(RandomAnchorOption{
	ID:               iotago.AddressFromArray([iotago.AddressLen]byte(testval.TestBytes(iotago.AddressLen, 1))),
	Assets:           &TestAssetsBag,
	AssetsReferentID: iotago.AddressFromArray([iotago.AddressLen]byte(testval.TestBytes(iotago.AddressLen, 2))),
	StateIndex:       lo.ToPtr[uint32](179537),
})
