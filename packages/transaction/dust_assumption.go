package transaction

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
)

type StorageDepositAssumption struct {
	AnchorOutput      uint64
	NativeTokenOutput uint64
	NFTOutput         uint64
}

func StorageDepositAssumptionFromBytes(data []byte) (*StorageDepositAssumption, error) {
	mu := marshalutil.New(data)
	var err error
	ret := &StorageDepositAssumption{}
	if ret.AnchorOutput, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.NativeTokenOutput, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.NFTOutput, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *StorageDepositAssumption) Bytes() []byte {
	return marshalutil.New().
		WriteUint64(d.AnchorOutput).
		WriteUint64(d.NativeTokenOutput).
		WriteUint64(d.NFTOutput).
		Bytes()
}

func (d *StorageDepositAssumption) String() string {
	return fmt.Sprintf("InternalStorageDepositEstimate: anchor UTXO = %d, nativetokenUTXO = %d",
		d.AnchorOutput, d.NativeTokenOutput)
}

func NewStorageDepositEstimate() *StorageDepositAssumption {
	return &StorageDepositAssumption{
		AnchorOutput:      aliasOutputStorageDeposit(),
		NativeTokenOutput: nativeTokenOutputStorageDeposit(),
		NFTOutput:         nftOutputStorageDeposit(),
	}
}

func aliasOutputStorageDeposit() uint64 {
	keyPair := cryptolib.NewKeyPairFromSeed([32]byte{})
	addr := keyPair.GetPublicKey().AsEd25519Address()

	aliasOutput := &iotago.AliasOutput{
		AliasID:       iotago.AliasID{},
		Amount:        1000,
		StateMetadata: state.OriginL1Commitment().Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: addr},
			&iotago.GovernorAddressUnlockCondition{Address: addr},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: addr,
			},
		},
	}
	return parameters.L1.Protocol.RentStructure.MinRent(aliasOutput)
}

func nativeTokenOutputStorageDeposit() uint64 {
	addr := iotago.AliasAddressFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, 0))
	o := MakeBasicOutput(
		&addr,
		&addr,
		&isc.FungibleTokens{
			BaseTokens: 1,
			Tokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     iotago.NativeTokenID{},
				Amount: abi.MaxUint256,
			}},
		},
		nil,
		isc.SendOptions{},
	)
	return parameters.L1.Protocol.RentStructure.MinRent(o)
}

func nftOutputStorageDeposit() uint64 {
	addr := iotago.AliasAddressFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, 0))
	basicOut := MakeBasicOutput(
		&addr,
		&addr,
		&isc.FungibleTokens{
			BaseTokens: 1,
			Tokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     iotago.NativeTokenID{},
				Amount: abi.MaxUint256,
			}},
		},
		nil,
		isc.SendOptions{},
	)
	out := NftOutputFromBasicOutput(basicOut, &isc.NFT{
		ID:       iotago.NFTID{0},
		Issuer:   tpkg.RandEd25519Address(),
		Metadata: make([]byte, iotago.MaxMetadataLength),
	})

	return parameters.L1.Protocol.RentStructure.MinRent(out)
}
