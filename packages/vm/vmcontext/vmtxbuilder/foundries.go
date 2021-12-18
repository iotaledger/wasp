package vmtxbuilder

import (
	"encoding/binary"
	"math/big"
	"sort"

	"github.com/iotaledger/hive.go/serializer/v2"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/util"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (txb *AnchorTransactionBuilder) CreateNewFoundry(
	scheme iotago.TokenScheme,
	tag iotago.TokenTag,
	maxSupply *big.Int,
) uint32 {
	f := &iotago.FoundryOutput{
		Address:           txb.anchorOutput.AliasID.ToAddress(),
		Amount:            0,
		NativeTokens:      nil,
		SerialNumber:      txb.nextFoundrySerialNumber(),
		TokenTag:          tag,
		CirculatingSupply: big.NewInt(0),
		MaximumSupply:     maxSupply,
		TokenScheme:       scheme,
		Blocks:            nil,
	}
	f.Amount = f.VByteCost(txb.rentStructure, nil)
	err := util.CatchPanicReturnError(func() {
		txb.subDeltaIotasFromTotal(f.Amount)
	}, ErrNotEnoughIotaBalance)
	if err != nil {
		panic(ErrNotEnoughFundsForInternalDustDeposit)
	}
	txb.invokedFoundries[f.SerialNumber] = &foundryInvoked{
		serialNumber: f.SerialNumber,
		in:           nil,
		out:          f,
	}
	return f.SerialNumber
}

func serNumFromNativeTokenID(tokenID *iotago.NativeTokenID) uint32 {
	slice := tokenID[iotago.AliasAddressSerializedBytesSize : iotago.AliasAddressSerializedBytesSize+serializer.UInt32ByteSize]
	return binary.LittleEndian.Uint32(slice)
}

// ModifyNativeTokenSupply inflates the supply is delta > 0, shrinks if delta < 0
func (txb *AnchorTransactionBuilder) ModifyNativeTokenSupply(tokenID *iotago.NativeTokenID, delta *big.Int) {
	serNum := serNumFromNativeTokenID(tokenID)
	nt, ok := txb.invokedFoundries[serNum]
	if !ok {
		// load foundry output from the state
		foundryOutput, inp := txb.loadFoundry(serNum)
		if foundryOutput == nil {
			panic(ErrFoundryDoesNotExist)
		}
		nt = &foundryInvoked{
			serialNumber: foundryOutput.SerialNumber,
			input:        *inp,
			in:           foundryOutput,
			out:          cloneFoundry(foundryOutput),
		}
	}
	// check if the loaded foundry matches the tokenID
	if *tokenID != nt.in.MustNativeTokenID() {
		panic(xerrors.Errorf("%w: requested token ID: %s, foundry token id: %s",
			ErrCantModifySupplyOfTheToken, tokenID.String(), nt.in.MustNativeTokenID().String()))
	}
	// check the supply bounds
	newSupply := big.NewInt(0).Add(nt.out.CirculatingSupply, delta)
	if newSupply.Cmp(big.NewInt(0)) < 0 || newSupply.Cmp(nt.out.MaximumSupply) > 0 {
		panic(ErrNativeTokenSupplyOutOffBounds)
	}
	// accrue/adjust this token balance in the internal outputs
	txb.addNativeTokenBalanceDelta(tokenID, delta)
	// update the supply and foundry record in the builder
	nt.out.CirculatingSupply = newSupply
	txb.invokedFoundries[serNum] = nt
}

func (txb *AnchorTransactionBuilder) nextFoundrySerialNumber() uint32 {
	numNew := uint32(0)
	for _, f := range txb.invokedFoundries {
		if f.isNewCreated() {
			numNew++
		}
	}
	return txb.anchorOutput.FoundryCounter + numNew
}

func (txb *AnchorTransactionBuilder) foundriesSorted() []*foundryInvoked {
	ret := make([]*foundryInvoked, 0, len(txb.invokedFoundries))
	for _, f := range txb.invokedFoundries {
		if !f.requiresInput() && !f.producesOutput() {
			continue
		}
		ret = append(ret, f)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].serialNumber < ret[j].serialNumber
	})
	return ret
}

func (f *foundryInvoked) clone() *foundryInvoked {
	return &foundryInvoked{
		in:  cloneFoundry(f.in),
		out: cloneFoundry(f.out),
	}
}

func (f *foundryInvoked) isNewCreated() bool {
	return !f.requiresInput() && f.producesOutput()
}

func (f *foundryInvoked) requiresInput() bool {
	if f.in == nil {
		return false
	}
	if identicalFoundries(f.in, f.out) {
		return false
	}
	return true
}

func (f *foundryInvoked) producesOutput() bool {
	if f.out == nil {
		return false
	}
	if identicalFoundries(f.in, f.out) {
		return false
	}
	return true
}

func cloneFoundry(f *iotago.FoundryOutput) *iotago.FoundryOutput {
	if f == nil {
		return nil
	}
	ret := &iotago.FoundryOutput{
		Address:           f.Address,
		Amount:            f.Amount,
		NativeTokens:      f.NativeTokens.Clone(),
		SerialNumber:      f.SerialNumber,
		TokenTag:          f.TokenTag,
		CirculatingSupply: f.CirculatingSupply,
		MaximumSupply:     f.MaximumSupply,
		TokenScheme:       f.TokenScheme,
		Blocks:            nil,
	}
	if !identicalFoundries(f, ret) {
		panic("cloneFoundry: very bad")
	}
	return ret
}

// identicalFoundries assumes use case and does consistency checks
func identicalFoundries(f1, f2 *iotago.FoundryOutput) bool {
	switch {
	case f1 == f2:
		return true
	case f1 == nil || f2 == nil:
		return false
	case f1.SerialNumber != f2.SerialNumber:
		return false
	case f1.CirculatingSupply.Cmp(f2.CirculatingSupply) != 0:
		return false
	case f1.Amount != f2.Amount:
		panic("identicalFoundries: inconsistency, amount is assumed immutable")
	case len(f1.NativeTokens) > 0 || len(f2.NativeTokens) > 0:
		panic("identicalFoundries: inconsistency, foundry is not expected not contain native tokens")
	case f1.MaximumSupply.Cmp(f2.MaximumSupply) != 0:
		panic("identicalFoundries: inconsistency, maximum supply is immutable")
	case !f1.Address.Equal(f2.Address):
		panic("identicalFoundries: inconsistency, addresses must always be equal")
	case f1.TokenScheme != f2.TokenScheme:
		panic("identicalFoundries: inconsistency, if serial numbers are equal, token schemes must be equal")
	case f1.TokenTag != f2.TokenTag:
		panic("identicalFoundries: inconsistency, if serial numbers are equal, token tags must be equal")
	case f1.Blocks != nil || f2.Blocks != nil:
		panic("identicalFoundries: inconsistency, feat blocks are not expected in the foundry")
	}
	return true
}
