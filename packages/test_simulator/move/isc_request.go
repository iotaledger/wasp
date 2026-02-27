package move

import (
	"fmt"

	"fortio.org/safecast"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

const reqAssetsBagSizeLimit = 25

// moveRequest mirrors the Move Request struct for BCS serialization.
// It uses Referent[AssetsBag] (not AssetsBagWithBalances) because the BCS-encoded
// Request from L1 does not include balance amounts — those are stored as dynamic
// fields on the bag object and must be fetched separately. This matches the
// intermediateMoveRequest pattern used in iscmoveclient/client_request.go.
type moveRequest struct {
	ID        iotago.ObjectID
	Sender    *cryptolib.Address
	AssetsBag iscmove.Referent[iscmove.AssetsBag]
	Message   iscmove.Message
	Allowance []byte
	GasBudget uint64
}

var RequestHandlers = map[string]l1.MoveCallFunc{
	"create_and_send_request": requestCreateAndSend,
	"destroy":                 requestDestroy,
	"receive":                 requestReceive,
}

// create_and_send_request(anchor, assets_bag, contract, function, args, allowance, gas_budget, ctx)
func requestCreateAndSend(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 7 {
		return nil, fmt.Errorf("request::create_and_send_request requires 7 arguments")
	}

	anchorAddr, err := extractAddress(args[0])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: anchor address: %w", err)
	}

	bag, err := extractBag(args[1])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: assets_bag: %w", err)
	}

	if bag.Size > reqAssetsBagSizeLimit {
		return nil, fmt.Errorf("request::create_and_send_request: assets bag size %d exceeds limit %d", bag.Size, reqAssetsBagSizeLimit)
	}

	contract, err := extractUint32(args[2])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: contract: %w", err)
	}

	function, err := extractUint32(args[3])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: function: %w", err)
	}

	msgArgs, err := extractByteVectors(args[4])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: args: %w", err)
	}

	allowance, err := extractBytes(args[5])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: allowance: %w", err)
	}

	gasBudget, err := extractUint64(args[6])
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: gas_budget: %w", err)
	}

	requestID := ctx.FreshID()
	assetsBagReferentID := ctx.FreshID()

	senderCrypto := cryptolib.NewAddressFromIota(&ctx.Sender)
	reqData := moveRequest{
		ID:     requestID,
		Sender: senderCrypto,
		AssetsBag: iscmove.Referent[iscmove.AssetsBag]{
			ID: assetsBagReferentID,
			Value: &iscmove.AssetsBag{
				ID:   bag.ID,
				Size: bag.Size,
			},
		},
		Message: iscmove.Message{
			Contract: contract,
			Function: function,
			Args:     msgArgs,
		},
		Allowance: allowance,
		GasBudget: gasBudget,
	}

	bcsData, err := bcs.Marshal(&reqData)
	if err != nil {
		return nil, fmt.Errorf("request::create_and_send_request: BCS marshal failed: %w", err)
	}

	reqObj := &l1.SimObject{
		ID:         requestID,
		Version:    0,
		Digest:     l1.ComputeDigest(bcsData),
		Owner:      l1.SimOwner{AddressOwner: &anchorAddr},
		Type:       l1.ISCTypeString(ctx.PackageID, iscmove.RequestModuleName, iscmove.RequestObjectName),
		Data:       bcsData,
		PreviousTx: ctx.TxDigest,
	}
	ctx.Store.Put(reqObj)

	return nil, nil
}

// destroy(request) -> (ID, AssetsBag)
func requestDestroy(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("request::destroy requires 1 argument")
	}

	reqID := args[0].ObjectID
	if reqID == nil {
		return nil, fmt.Errorf("request::destroy: argument has no object ID")
	}

	reqObj, ok := ctx.Store.Get(*reqID)
	if !ok {
		return nil, fmt.Errorf("request::destroy: request %s not found", reqID.String())
	}

	req, err := bcs.Unmarshal[moveRequest](reqObj.Data)
	if err != nil {
		return nil, fmt.Errorf("request::destroy: BCS unmarshal failed: %w", err)
	}

	ctx.Store.Delete(*reqID)

	bag := &AssetsBagValue{
		ID:   req.AssetsBag.Value.ID,
		Size: req.AssetsBag.Value.Size,
	}

	return []l1.Value{
		{ObjectID: reqID, Type: "ID"},
		{ObjectID: &bag.ID, Raw: bag, Type: "AssetsBag"},
	}, nil
}

// receive(parent_uid, receiving) -> Request
func requestReceive(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	return transferReceive(ctx, nil, args)
}

func extractAddress(v l1.Value) (iotago.Address, error) {
	switch val := v.Raw.(type) {
	case *iotago.Address:
		return *val, nil
	case iotago.Address:
		return val, nil
	case []byte:
		// BCS-encoded address: 32 raw bytes
		if len(val) == 32 {
			var addr iotago.Address
			copy(addr[:], val)
			return addr, nil
		}
		return iotago.Address{}, fmt.Errorf("expected 32-byte Address, got %d bytes", len(val))
	default:
		return iotago.Address{}, fmt.Errorf("expected Address, got %T", v.Raw)
	}
}

func extractUint32(v l1.Value) (uint32, error) {
	switch val := v.Raw.(type) {
	case uint32:
		return val, nil
	case uint64:
		return safecast.Convert[uint32](val)
	case int:
		return safecast.Convert[uint32](val)
	case []byte:
		result, err := bcs.Unmarshal[uint32](val)
		if err != nil {
			return 0, fmt.Errorf("BCS decode u32 failed: %w", err)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("expected uint32, got %T", v.Raw)
	}
}

func extractBytes(v l1.Value) ([]byte, error) {
	switch val := v.Raw.(type) {
	case []byte:
		if decoded, err := bcs.Unmarshal[[]byte](val); err == nil {
			return decoded, nil
		}
		return val, nil
	case *[]byte:
		if val == nil {
			return nil, nil
		}
		return *val, nil
	default:
		return nil, fmt.Errorf("expected []byte, got %T", v.Raw)
	}
}

func extractByteVectors(v l1.Value) ([][]byte, error) {
	switch val := v.Raw.(type) {
	case [][]byte:
		return val, nil
	case []l1.Value:
		result := make([][]byte, len(val))
		for i, item := range val {
			b, err := extractBytes(item)
			if err != nil {
				return nil, err
			}
			result[i] = b
		}
		return result, nil
	case []byte:
		result, err := bcs.Unmarshal[[][]byte](val)
		if err != nil {
			return nil, fmt.Errorf("BCS decode vector<vector<u8>> failed: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected [][]byte, got %T", v.Raw)
	}
}
