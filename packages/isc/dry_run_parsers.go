package isc

import (
	"errors"
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type EstimationRequest struct {
	Message      iscmove.Message
	AllowanceBCS []byte
	GasBudget    uint64
}

// getPureInput extracts the Pure bytes from a CallArg referenced by an Argument with Input type.
func getPureInput(inputs []iotago.CallArg, arg iotago.Argument) ([]byte, error) {
	if arg.Input == nil {
		return nil, fmt.Errorf("expected Input argument, got %s", arg.String())
	}
	idx := int(*arg.Input)
	if idx >= len(inputs) {
		return nil, fmt.Errorf("input index %d out of range (have %d inputs)", idx, len(inputs))
	}
	if inputs[idx].Pure == nil {
		return nil, fmt.Errorf("expected Pure input at index %d", idx)
	}
	return *inputs[idx].Pure, nil
}

func DecodeCreateAndSendRequest(msg *EstimationRequest, cmd *iotago.ProgrammableMoveCall, inputs []iotago.CallArg) error {
	if len(cmd.Arguments) != 7 {
		return errors.New("create_and_send_request has invalid parameters")
	}

	// contractHname
	pureBytes, err := getPureInput(inputs, cmd.Arguments[2])
	if err != nil {
		return fmt.Errorf("failed to get contract hname input: %w", err)
	}
	msg.Message.Contract, err = bcs.Unmarshal[uint32](pureBytes)
	if err != nil {
		return fmt.Errorf("failed to decode contract hname: %w", err)
	}

	// functionHname
	pureBytes, err = getPureInput(inputs, cmd.Arguments[3])
	if err != nil {
		return fmt.Errorf("failed to get function hname input: %w", err)
	}
	msg.Message.Function, err = bcs.Unmarshal[uint32](pureBytes)
	if err != nil {
		return fmt.Errorf("failed to decode function hname: %w", err)
	}

	// contractCallArgs
	pureBytes, err = getPureInput(inputs, cmd.Arguments[4])
	if err != nil {
		return fmt.Errorf("failed to get contract call args input: %w", err)
	}
	msg.Message.Args, err = bcs.Unmarshal[[][]byte](pureBytes)
	if err != nil {
		return fmt.Errorf("failed to decode contract call args: %w", err)
	}

	// allowance
	pureBytes, err = getPureInput(inputs, cmd.Arguments[5])
	if err != nil {
		return fmt.Errorf("failed to get allowance input: %w", err)
	}
	msg.AllowanceBCS, err = bcs.Unmarshal[[]byte](pureBytes)
	if err != nil {
		return fmt.Errorf("failed to decode allowance: %w", err)
	}

	// gasBudget
	pureBytes, err = getPureInput(inputs, cmd.Arguments[6])
	if err != nil {
		return fmt.Errorf("failed to get gas budget input: %w", err)
	}
	msg.GasBudget, err = bcs.Unmarshal[uint64](pureBytes)
	if err != nil {
		return fmt.Errorf("failed to decode gas budget: %w", err)
	}

	return nil
}

func DecodeCoin(assets *Assets, cmd *iotago.ProgrammableMoveCall, allCommands []iotago.Command, inputs []iotago.CallArg) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("malformed PTB: place_coin expects 2 arguments, got %d", len(cmd.Arguments))
	}

	// place_coin arguments: [assetsBag, coin]
	// The coin (arg[1]) is a Result from a SplitCoins command.
	// We trace back to the SplitCoins to extract the amount from its Pure input.
	coinArg := cmd.Arguments[1]
	if coinArg.Result == nil {
		return fmt.Errorf("expected Result argument for coin in place_coin")
	}

	cmdIdx := int(*coinArg.Result)
	if cmdIdx >= len(allCommands) {
		return fmt.Errorf("command index %d out of range", cmdIdx)
	}

	splitCmd := allCommands[cmdIdx]
	if splitCmd.SplitCoins == nil {
		return fmt.Errorf("expected SplitCoins command producing coin for place_coin, got command at index %d", cmdIdx)
	}
	if len(splitCmd.SplitCoins.Amounts) == 0 {
		return fmt.Errorf("SplitCoins has no amounts")
	}

	amountBytes, err := getPureInput(inputs, splitCmd.SplitCoins.Amounts[0])
	if err != nil {
		return fmt.Errorf("can't get amount from SplitCoins: %w", err)
	}

	amount, err := bcs.Unmarshal[uint64](amountBytes)
	if err != nil {
		return fmt.Errorf("can't decode amount: %w", err)
	}

	assets.AddCoin(coin.MustTypeFromString(cmd.TypeArguments[0].String()), coin.Value(amount))
	return nil
}

func DecodeAsset(assets *Assets, cmd *iotago.ProgrammableMoveCall) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("malformed PTB")
	}

	if len(cmd.TypeArguments) != 1 {
		return fmt.Errorf("malformed PTB")
	}

	objectType, err := iotago.ObjectTypeFromString(cmd.TypeArguments[0].String())
	if err != nil {
		return fmt.Errorf("can't decode typeTag in place_asset command: %v", err)
	}

	// Not parsing objectID here, because it could be a reference to the result of other command.

	assets.AddObject(IotaObject{
		ID:   *iotatest.RandomAddress(),
		Type: objectType,
	})

	return err
}

// DecodeDryRunTransaction decodes the transaction from a dry run result to extract
// assets, request info, and sender address.
// TODO: This needs proper implementation - currently decodes the BCS transaction from the dry run response.
func DecodeDryRunTransaction(dryRunRes *graphqltypes.DryRunTransactionBlockDryRunTransactionBlockDryRunResult) (*Assets, *EstimationRequest, *cryptolib.Address, error) {
	txBcs := dryRunRes.Transaction.Bcs
	txData, err := bcs.Unmarshal[iotago.TransactionData](txBcs)
	if err != nil {
		return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("failed to unmarshal transaction BCS: %w", err)
	}
	if txData.V1 == nil {
		return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("only TransactionData V1 is supported")
	}
	pt := txData.V1.Kind.ProgrammableTransaction
	if pt == nil {
		return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("transaction is not a ProgrammableTransaction")
	}

	assets := NewAssets(0)
	request := &EstimationRequest{
		Message: iscmove.Message{},
	}

	for _, command := range pt.Commands {
		if cmd := command.MoveCall; cmd != nil {
			if cmd.Function == "place_coin" {
				if err := DecodeCoin(assets, cmd, pt.Commands, pt.Inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_coin command: %w", err)
				}
			}

			if cmd.Function == "place_asset" {
				if err := DecodeAsset(assets, cmd); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_asset command: %w", err)
				}
			}

			if cmd.Function == "create_and_send_request" {
				if err := DecodeCreateAndSendRequest(request, cmd, pt.Inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode create_and_send_request command: %w", err)
				}
			}
		}
	}

	return assets, request, cryptolib.NewAddressFromIota(&txData.V1.Sender), nil
}
