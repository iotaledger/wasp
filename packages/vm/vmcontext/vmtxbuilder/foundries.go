package vmtxbuilder

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/util"

	iotago "github.com/iotaledger/iota.go/v3"
)

func (txb *AnchorTransactionBuilder) CreateNewFoundry(
	scheme iotago.TokenScheme,
	tag iotago.TokenTag,
	maxSupply *big.Int,
	metadata []byte,
) (uint32, uint64) {
	if maxSupply.Cmp(big.NewInt(0)) <= 0 {
		panic(ErrCreateFoundryMaxSupplyMustBePositive)
	}
	if maxSupply.Cmp(abi.MaxUint256) > 0 {
		panic(ErrCreateFoundryMaxSupplyTooBig)
	}

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
	if len(metadata) > 0 {
		f.Blocks = iotago.FeatureBlocks{&iotago.MetadataFeatureBlock{
			Data: metadata,
		}}
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
	return f.SerialNumber, f.Amount
}

// ModifyNativeTokenSupply inflates the supply is delta > 0, shrinks if delta < 0
// returns adjustment of the dust deposit.
func (txb *AnchorTransactionBuilder) ModifyNativeTokenSupply(tokenID *iotago.NativeTokenID, delta *big.Int) int64 {
	txb.MustBalanced("ModifyNativeTokenSupply: IN")
	sn := accounts.FoundrySNFromNativeTokenID(tokenID)
	nt, ok := txb.invokedFoundries[sn]
	if !ok {
		// load foundry output from the state
		foundryOutput, inp := txb.loadFoundry(sn)
		if foundryOutput == nil {
			panic(ErrFoundryDoesNotExist)
		}
		nt = &foundryInvoked{
			serialNumber: foundryOutput.SerialNumber,
			input:        *inp,
			in:           foundryOutput,
			out:          cloneFoundryOutput(foundryOutput),
		}
	}
	// check if the loaded foundry matches the tokenID
	if *tokenID != nt.in.MustNativeTokenID() {
		panic(xerrors.Errorf("%v: requested token ID: %s, foundry token id: %s",
			ErrCantModifySupplyOfTheToken, tokenID.String(), nt.in.MustNativeTokenID().String()))
	}
	// check the supply bounds
	newSupply := big.NewInt(0).Add(nt.out.CirculatingSupply, delta)
	if newSupply.Cmp(big.NewInt(0)) < 0 || newSupply.Cmp(nt.out.MaximumSupply) > 0 {
		panic(ErrNativeTokenSupplyOutOffBounds)
	}
	// accrue/adjust this token balance in the internal outputs
	adjustment := txb.addNativeTokenBalanceDelta(tokenID, delta)
	// update the supply and foundry record in the builder
	nt.out.CirculatingSupply = newSupply
	txb.invokedFoundries[sn] = nt

	adjustment += int64(nt.in.Amount) - int64(nt.out.Amount)
	txb.MustBalanced("ModifyNativeTokenSupply: OUT")
	return adjustment
}

func (txb *AnchorTransactionBuilder) nextFoundrySerialNumber() uint32 {
	return txb.nextFoundryCounter() + 1
}

func (txb *AnchorTransactionBuilder) nextFoundryCounter() uint32 {
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

func (txb *AnchorTransactionBuilder) FoundriesToBeUpdated() ([]uint32, []uint32) {
	toBeUpdated := make([]uint32, 0, len(txb.invokedFoundries))
	toBeRemoved := make([]uint32, 0, len(txb.invokedFoundries))
	for _, f := range txb.foundriesSorted() {
		if f.producesOutput() {
			toBeUpdated = append(toBeUpdated, f.serialNumber)
		} else if f.requiresInput() {
			toBeRemoved = append(toBeRemoved, f.serialNumber)
		}
	}
	return toBeUpdated, toBeRemoved
}

func (txb *AnchorTransactionBuilder) FoundryOutputsBySN(serNums []uint32) map[uint32]*iotago.FoundryOutput {
	ret := make(map[uint32]*iotago.FoundryOutput)
	for _, sn := range serNums {
		ret[sn] = txb.invokedFoundries[sn].out
	}
	return ret
}

func (f *foundryInvoked) clone() *foundryInvoked {
	return &foundryInvoked{
		in:  cloneFoundryOutput(f.in),
		out: cloneFoundryOutput(f.out),
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

func cloneFoundryOutput(f *iotago.FoundryOutput) *iotago.FoundryOutput {
	if f == nil {
		return nil
	}
	ret := &iotago.FoundryOutput{
		Address:           f.Address,
		Amount:            f.Amount,
		NativeTokens:      f.NativeTokens.Clone(),
		SerialNumber:      f.SerialNumber,
		TokenTag:          f.TokenTag,
		CirculatingSupply: new(big.Int).Set(f.CirculatingSupply),
		MaximumSupply:     new(big.Int).Set(f.MaximumSupply),
		TokenScheme:       f.TokenScheme,
		Blocks:            nil,
	}
	if !identicalFoundries(f, ret) {
		panic("cloneFoundryOutput: very bad")
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
