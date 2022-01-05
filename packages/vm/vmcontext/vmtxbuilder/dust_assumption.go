package vmtxbuilder

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

type InternalDustDepositAssumption struct {
	AnchorOutput      uint64
	NativeTokenOutput uint64
}

func InternalDustDepositAssumptionFromBytes(data []byte) (*InternalDustDepositAssumption, error) {
	mu := marshalutil.New(data)
	var err error
	ret := &InternalDustDepositAssumption{}
	if ret.AnchorOutput, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.NativeTokenOutput, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *InternalDustDepositAssumption) Bytes() []byte {
	return marshalutil.New().
		WriteUint64(d.AnchorOutput).
		WriteUint64(d.NativeTokenOutput).
		Bytes()
}

func (d *InternalDustDepositAssumption) String() string {
	return fmt.Sprintf("InternalDustDepositEstimate: anchor UTXO = %d, nativetokenUTXO = %d",
		d.AnchorOutput, d.NativeTokenOutput)
}

func NewDepositEstimate(rent *iotago.RentStructure) *InternalDustDepositAssumption {
	return &InternalDustDepositAssumption{
		AnchorOutput:      aliasOutputDustDeposit(rent) + 50,
		NativeTokenOutput: nativeTokenOutputDustDeposit(rent) + 50,
	}
}

func aliasOutputDustDeposit(rent *iotago.RentStructure) uint64 {
	keyPair := cryptolib.NewKeyPairFromSeed([32]byte{})
	addr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)

	aliasOutput := &iotago.AliasOutput{
		AliasID:              iotago.AliasID{},
		Amount:               1000,
		StateController:      addr,
		GovernanceController: addr,
		StateMetadata:        state.OriginStateHash().Bytes(),
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: addr,
			},
		},
	}
	return aliasOutput.VByteCost(rent, nil)
}

func nativeTokenOutputDustDeposit(rent *iotago.RentStructure) uint64 {
	addr := iotago.AliasAddressFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, 0))
	o := MakeExtendedOutput(
		&addr,
		&addr,
		&iscp.Assets{
			Iotas: 1,
			Tokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     iotago.NativeTokenID{},
				Amount: abi.MaxUint256,
			}},
		},
		nil,
		nil,
		rent,
	)
	return o.VByteCost(rent, nil)
}
