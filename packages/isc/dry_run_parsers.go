package isc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type EstimationRequest struct {
	Message      iscmove.Message
	AllowanceBCS []byte
	GasBudget    json.Number
}

func DecodeCreateAndSendRequest(msg *EstimationRequest, cmd *iotago.ProgrammableMoveCall, inputs []iotagraphql.ProgrammableTransactionBlockPureInput) error {
	if len(cmd.Arguments) != 7 {
		return errors.New("create_and_send_request has invalid parameters")
	}

	// contractHname
	err := json.Unmarshal(inputs[*cmd.Arguments[2].Input].Value, &msg.Message.Contract)
	if err != nil {
		return fmt.Errorf("failed to decode contract hname: %v", err)
	}

	// functionHname
	err = json.Unmarshal(inputs[*cmd.Arguments[3].Input].Value, &msg.Message.Function)
	if err != nil {
		return fmt.Errorf("failed to decode function hname: %v", err)
	}

	// contractCallArgs
	err = json.Unmarshal(inputs[*cmd.Arguments[4].Input].Value, &msg.Message.Args)
	if err != nil {
		return fmt.Errorf("failed to decode contract call args: %v", err)
	}

	// allowance
	err = json.Unmarshal(inputs[*cmd.Arguments[5].Input].Value, &msg.AllowanceBCS)
	if err != nil {
		return fmt.Errorf("failed to decode allowance: %v", err)
	}

	// gasBudget
	err = json.Unmarshal(inputs[*cmd.Arguments[6].Input].Value, &msg.GasBudget)
	if err != nil {
		return fmt.Errorf("failed to decode gas budget: %v", err)
	}

	return nil
}

func DecodeCoin(assets *Assets, cmd *iotago.ProgrammableMoveCall, inputs []iotagraphql.ProgrammableTransactionBlockPureInput) error {
	var err error
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("malformed PTB")
	}

	var amountString string
	err = json.Unmarshal(inputs[*cmd.Arguments[0].Result].Value, &amountString)
	if err != nil {
		return fmt.Errorf("malformed PTB")
	}

	amount, err := strconv.ParseUint(amountString, 10, 64)
	if err != nil {
		err = fmt.Errorf("can't decode amount argument in place_coin command: %w", err)
		return err
	}

	assets.AddCoin(coin.MustTypeFromString(cmd.TypeArguments[0].String()), coin.Value(amount))
	return nil
}

func DecodeAsset(assets *Assets, cmd *iotago.ProgrammableMoveCall, inputs []iotagraphql.ProgrammableTransactionBlockPureInput) error {
	var err error
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

// DecodeDryRunTransaction The intention of this parser is to make the use of the gas estimation easier.
// We only accept the transactionBytes and select all needed inputs.
// The upside is that a user can pass an unsigned transaction to estimate.
// The downside is that any time we change create_and_send_request in the move contract, we need to update this logic.
// I don't expect it to change often if ever, so that seems to be a straight forward way.
func DecodeDryRunTransaction(dryRunRes *iotagraphql.DryRunResult) (*Assets, *EstimationRequest, *cryptolib.Address, error) {
	tx := dryRunRes.Input.Data.V1.Transaction.Data.ProgrammableTransaction

	var cmds []struct {
		MoveCall *iotago.ProgrammableMoveCall `json:"MoveCall,omitempty"`
	}
	if err := json.Unmarshal(tx.Commands, &cmds); err != nil {
		return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode dry run response: %w", err)
	}

	assets := NewAssets(0)
	request := &EstimationRequest{
		Message: iscmove.Message{},
	}

	for _, moveCall := range cmds {
		if cmd := moveCall.MoveCall; cmd != nil {
			// take all placed coins into assets
			if cmd.Function == "place_coin" {
				var inputs []iotagraphql.ProgrammableTransactionBlockPureInput
				if err := json.Unmarshal(tx.Inputs, &inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_coin command: %w", err)
				}

				if err := DecodeCoin(assets, cmd, inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_coin command: %w", err)
				}
			}

			if cmd.Function == "place_asset" {
				var inputs []iotagraphql.ProgrammableTransactionBlockPureInput
				if err := json.Unmarshal(tx.Inputs, &inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_asset command: %w", err)
				}

				if err := DecodeAsset(assets, cmd, inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode place_asset command: %w", err)
				}
			}

			if cmd.Function == "create_and_send_request" {
				var inputs []iotagraphql.ProgrammableTransactionBlockPureInput
				if err := json.Unmarshal(tx.Inputs, &inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode create_and_send_request command: %w", err)
				}

				if err := DecodeCreateAndSendRequest(request, cmd, inputs); err != nil {
					return nil, nil, cryptolib.NewEmptyAddress(), fmt.Errorf("can't decode create_and_send_request command: %w", err)
				}
			}
		}
	}

	return assets, request, cryptolib.NewAddressFromIota(&dryRunRes.Input.Data.V1.Sender), nil
}
