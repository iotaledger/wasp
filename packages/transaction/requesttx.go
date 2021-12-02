package transaction

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

type NewRequestTransactionParams struct {
	SenderKeyPair    *ed25519.KeyPair
	UnspentOutputs   []iotago.Output
	UnspentOutputIDs []*iotago.UTXOInput
	Requests         []*iscp.RequestParameters
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
// To avoid empty transfer it defaults to 1 iota
func NewRequestTransaction(par NewRequestTransactionParams) (*iotago.Transaction, error) {
	outputs := iotago.Outputs{}
	sumIotasOut := uint64(0)
	sumTokensOut := make(map[iotago.NativeTokenID]*big.Int)

	var senderAddress iotago.Address // TODO get address from par.SenderKeyPair.PublicKey.

	// create outputs, sum totals needed
	for _, req := range par.Requests {
		out, _ := vmtxbuilder.NewExtendedOutput(
			req.Target,
			req.Assets,
			senderAddress,
			&iscp.RequestMetadata{
				SenderContract: 0,
				TargetContract: req.Metadata.TargetContract,
				EntryPoint:     req.Metadata.EntryPoint,
				Params:         req.Metadata.Params,
				Transfer:       req.Metadata.Transfer,
				GasBudget:      req.Metadata.GasBudget,
			},
			req.Options,
		)
		outputs = append(outputs, out)
		sumIotasOut += out.Amount
		for _, nt := range req.Assets.Tokens {
			s, ok := sumTokensOut[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokensOut[nt.ID] = s
		}
	}
	sumIotasIn := uint64(0)
	sumTokensIn := make(map[iotago.NativeTokenID]*big.Int)

	var remIotas uint64
	var remTokens map[iotago.NativeTokenID]*big.Int
	var enough bool
	var inputs iotago.Inputs
	for i, inp := range par.UnspentOutputs {
		a := vmtxbuilder.AssetsFromOutput(inp)
		sumIotasIn += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := sumTokensIn[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokensIn[nt.ID] = s
		}
		// calculate reminders
		remIotas, remTokens, enough = calcReminders(sumIotasIn, sumIotasOut, sumTokensIn, sumTokensOut)
		if enough {
			for _, input := range par.UnspentOutputIDs[:i] {
				inputs = append(inputs, input)
			}
			break
		}
	}
	if !enough {
		return nil, accounts.ErrNotEnoughFunds
	}
	// enough funds, create reminder output if needed
	if remIotas > 0 {
		a := &iscp.Assets{
			Iotas:  remIotas,
			Tokens: iotago.NativeTokens{},
		}
		for id, b := range remTokens {
			a.Tokens = append(a.Tokens, &iotago.NativeToken{
				ID:     id,
				Amount: b,
			})
		}
		reminderOutput := &iotago.ExtendedOutput{
			Address:      senderAddress,
			Amount:       remIotas,
			NativeTokens: a.Tokens,
			Blocks:       nil,
		}
		outputs = append(outputs, reminderOutput)
	}
	essence := iotago.TransactionEssence{
		Inputs:  inputs,
		Outputs: outputs,
		Payload: nil,
	}
	// TODO sign the transaction essence and create unlock blocks
	// essence.Sign()
}

func calcReminders(inIotas, outIotas uint64, inTokens, outTokens map[iotago.NativeTokenID]*big.Int) (uint64, map[iotago.NativeTokenID]*big.Int, bool) {
	if inIotas < outIotas {
		return 0, nil, false
	}
	retIotas := outIotas - inIotas
	retTokens := make(map[iotago.NativeTokenID]*big.Int)

	for id, bOut := range outTokens {
		bIn, ok := inTokens[id]
		if !ok {
			return 0, nil, false
		}
		if bIn.Cmp(bOut) < 0 {
			return 0, nil, false
		}
		s := new(big.Int).Sub(bIn, bOut)
		if !util.IsZeroBigInt(bIn) {
			retTokens[id] = s
		}
	}
	if len(retTokens) > 0 && retIotas < calcDustDepositByNativeTokenBalances(retTokens) {
		return 0, nil, false
	}
	return retIotas, retTokens, true
}

func calcDustDepositByNativeTokenBalances(tokens map[iotago.NativeTokenID]*big.Int) uint64 {
	return parameters.DeSerializationParameters().MinDustDeposit // TODO take into account number of native tokens
}
