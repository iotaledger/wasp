package isc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/tidwall/gjson"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

type OnLedgerRequestData struct {
	requestRef      iotago.ObjectRef               `bcs:"export"`
	senderAddress   *cryptolib.Address             `bcs:"export"`
	targetAddress   *cryptolib.Address             `bcs:"export"`
	assets          *Assets                        `bcs:"export"`
	assetsBag       *iscmove.AssetsBagWithBalances `bcs:"export"`
	requestMetadata *RequestMetadata               `bcs:"export"`
}

var (
	_ Request         = new(OnLedgerRequestData)
	_ OnLedgerRequest = new(OnLedgerRequestData)
	_ Calldata        = new(OnLedgerRequestData)
)

func OnLedgerFromMoveRequest(request *iscmove.RefWithObject[iscmove.Request], anchorAddress *cryptolib.Address) (OnLedgerRequest, error) {
	assets, err := AssetsFromAssetsBagWithBalances(&request.Object.AssetsBag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assets from AssetsBag: %w", err)
	}
	return &OnLedgerRequestData{
		requestRef:    request.ObjectRef,
		senderAddress: request.Object.Sender,
		targetAddress: anchorAddress,
		assets:        assets,
		assetsBag:     &request.Object.AssetsBag,
		requestMetadata: &RequestMetadata{
			SenderContract: ContractIdentity{},
			Message: Message{
				Target: CallTarget{
					Contract:   Hname(request.Object.Message.Contract),
					EntryPoint: Hname(request.Object.Message.Function),
				},
				Params: request.Object.Message.Args,
			},
			AllowanceBCS: request.Object.AllowanceBCS,
			GasBudget:    request.Object.GasBudget,
		},
	}, nil
}

func FakeEstimateOnLedger(dryRunRes *iotajsonrpc.DryRunTransactionBlockResponse, msg *iscmove.Request) (OnLedgerRequest, error) {
	assets := NewAssets(0)
	allowance := NewAssets(0)
	tx := dryRunRes.Input.Data.V1.Transaction.Data.ProgrammableTransaction
	cmds := gjson.ParseBytes(tx.Commands)
	var err error
	cmds.ForEach(func(key, value gjson.Result) bool {
		if moveCall := value.Get("MoveCall"); moveCall.Exists() {
			var cmd iotago.ProgrammableMoveCall
			err = json.Unmarshal([]byte(moveCall.String()), &cmd)
			if err != nil {
				err = fmt.Errorf("can't decode dry run response: %w", err)
				return false
			}

			// take all placed coins into assets
			if cmd.Function == "place_coin" {
				var inputs []iotajsonrpc.ProgrammableTransactionBlockPureInput
				err = json.Unmarshal(tx.Inputs, &inputs)
				if err != nil {
					err = fmt.Errorf("can't decode place_coin command: %w", err)
					return false
				}

				if len(cmd.Arguments) < 1 {
					err = fmt.Errorf("malformed PTB")
					return false
				}

				var amountString string
				err = json.Unmarshal(inputs[*cmd.Arguments[0].Result].Value, &amountString)
				if err != nil {
					err = fmt.Errorf("malformed PTB")
					return false
				}

				amount, err := strconv.ParseUint(amountString, 10, 64)
				if err != nil {
					err = fmt.Errorf("can't decode value in place_coin command: %w", err)
					return false
				}
				assets.AddCoin(coin.MustTypeFromString(cmd.TypeArguments[0].String()), coin.Value(amount))
			}

			if cmd.Function == "create_and_send_request" {
				var inputs []iotajsonrpc.ProgrammableTransactionBlockPureInput
				err = json.Unmarshal(tx.Inputs, &inputs)
				if err != nil {
					err = fmt.Errorf("can't decode place_coin command: %w", err)
					return false
				}

				if len(cmd.Arguments) < 7 {
					err = fmt.Errorf("malformed PTB")
					return false
				}

				rawAllowanceBalance, err := json.Marshal(inputs[*cmd.Arguments[6].Input])
				if err != nil {
					err = fmt.Errorf("can't decode allowance of create_and_send_request command: %w", err)
					return false
				}
				var allowanceInputRaw iotajsonrpc.ProgrammableTransactionBlockPureInput
				err = json.Unmarshal(rawAllowanceBalance, &allowanceInputRaw)
				if err != nil {
					err = fmt.Errorf("can't decode allowance of create_and_send_request command: %w", err)
					return false
				}

				var allowanceRaw []byte
				err = json.Unmarshal(allowanceInputRaw.Value, &allowanceRaw)
				if err != nil {
					err = fmt.Errorf("can't decode allowance of create_and_send_request command: %w", err)
					return false
				}
				allowanceTmp, err := bcs.Unmarshal[Assets](allowanceRaw)
				if err != nil {
					err = fmt.Errorf("can't decode allowance of create_and_send_request command: %w", err)
					return false
				}
				*allowance = allowanceTmp
			}
		}
		return true // Continue iteration
	})
	if err != nil {
		return nil, err
	}

	allowanceAsAssets, err := AssetsFromISCMove(&msg.Allowance)
	if err != nil {
		return nil, err
	}

	r := &OnLedgerRequestData{
		requestRef:    *iotatest.RandomObjectRef(),
		senderAddress: cryptolib.NewRandomAddress(),
		targetAddress: cryptolib.NewRandomAddress(),
		assets:        assets,
		assetsBag:     &iscmove.AssetsBagWithBalances{},
		requestMetadata: &RequestMetadata{
			SenderContract: ContractIdentity{},
			Message: Message{
				Target: CallTarget{
					Contract:   Hname(msg.Contract),
					EntryPoint: Hname(msg.Function),
				},
				Params: msg.Args,
			},
			Allowance: allowanceAsAssets,
			GasBudget: iotaclient.DefaultGasBudget,
		},
	}
	return r, nil
}

func (req *OnLedgerRequestData) Allowance() (*Assets, error) {
	if req.requestMetadata == nil || len(req.requestMetadata.AllowanceBCS) == 0 {
		return NewEmptyAssets(), nil
	}
	assets, err := bcs.Unmarshal[iscmove.Assets](req.requestMetadata.AllowanceBCS)
	if err != nil {
		return nil, err
	}
	return AssetsFromISCMove(&assets)
}

func (req *OnLedgerRequestData) Assets() *Assets {
	return req.assets
}

func (req *OnLedgerRequestData) Bytes() []byte {
	var r Request = req
	return bcs.MustMarshal(&r)
}

func (req *OnLedgerRequestData) Message() Message {
	if req.requestMetadata == nil {
		return Message{}
	}
	return req.requestMetadata.Message
}

func (req *OnLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	if req.requestMetadata == nil {
		return 0, false
	}
	return req.requestMetadata.GasBudget, false
}

func (req *OnLedgerRequestData) ID() RequestID {
	return RequestID(*req.requestRef.ObjectID)
}

func (req *OnLedgerRequestData) IsOffLedger() bool {
	return false
}

func (req *OnLedgerRequestData) RequestID() iotago.ObjectID {
	return *req.requestRef.ObjectID
}

func (req *OnLedgerRequestData) SenderAccount() AgentID {
	sender := req.SenderAddress()
	if sender == nil {
		return nil
	}
	if req.requestMetadata != nil && !req.requestMetadata.SenderContract.Empty() {
		chainID := ChainIDFromAddress(sender)
		return req.requestMetadata.SenderContract.AgentID(chainID)
	}
	return NewAddressAgentID(sender)
}

func (req *OnLedgerRequestData) SenderAddress() *cryptolib.Address {
	return req.senderAddress
}

func (req *OnLedgerRequestData) String() string {
	metadata := req.requestMetadata
	if metadata == nil {
		return "onledger request without metadata"
	}

	return fmt.Sprintf("OnLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, Assets: %v, GasBudget: %d }",
		req.ID().String(),
		req.SenderAddress().String(),
		metadata.Message.Target.Contract.String(),
		metadata.Message.Target.EntryPoint.String(),
		metadata.Message.Params.String(),
		req.assets,
		metadata.GasBudget,
	)
}

func (req *OnLedgerRequestData) RequestRef() iotago.ObjectRef {
	return req.requestRef
}

func (req *OnLedgerRequestData) AssetsBag() *iscmove.AssetsBagWithBalances {
	return req.assetsBag
}

func (req *OnLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}
