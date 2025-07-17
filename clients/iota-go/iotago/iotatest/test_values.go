package iotatest

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/samber/lo"
)

var TestTransactionData = testTransactionData(
	lo.Must(iotago.AddressFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")),
	lo.Must(iotago.AddressFromHex("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd")),
)

func testTransactionData(packageAddr, senderAddr *iotago.Address) *iotago.TransactionData {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       packageAddr,
				Module:        "test_module",
				Function:      "test_func",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		senderAddr,
		pt,
		[]*iotago.ObjectRef{},
		10000,
		100,
	)
	return &tx
}

var TestAddress = iotago.AddressFromArray([iotago.AddressLen]byte(testutil.TestBytes(iotago.AddressLen)))
var TestDigest = iotago.MustNewDigest(testutil.TestHex(iotago.DigestSize))

var TestObjectRef = &iotago.ObjectRef{
	ObjectID: TestAddress,
	Version:  719,
	Digest:   TestDigest,
}
