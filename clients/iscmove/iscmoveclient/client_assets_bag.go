package iscmoveclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
)

func (c *Client) GetAssetsBagWithBalances(
	ctx context.Context,
	assetsBagID *iotago.ObjectID,
) (*iscmove.AssetsBagWithBalances, error) {
	fields, err := c.GetDynamicFields(ctx, iotagraphql.GetDynamicFieldsRequest{ParentObjectID: *assetsBagID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in AssetsBag: %w", err)
	}

	nodes := fields.Owner.DynamicFields.Nodes
	bag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *assetsBagID,
			Size: uint64(len(nodes)),
		},
		Assets: *iscmove.NewEmptyAssets(),
	}
	for _, node := range nodes {
		nameTypeRepr := node.Name.Type.Repr
		isCoin, err := iotago.IsSameResource(nameTypeRepr, "0x1::ascii::String")
		if err != nil {
			return nil, fmt.Errorf("failed to check if resource is coin: %w", err)
		}

		if isCoin {
			// Extract the coin type string from the JSON name value
			var nameStr string
			if err := json.Unmarshal(node.Name.Json, &nameStr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal coin type name: %w", err)
			}
			cointype, err := iotagraphql.CoinTypeFromString("0x" + nameStr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert cointype: %w", err)
			}

			var balanceJSON json.RawMessage
			switch v := node.Value.(type) {
			case *graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject:
				balanceJSON = v.Contents.Json
			case *graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveValue:
				balanceJSON = v.Json
			default:
				return nil, fmt.Errorf("coin dynamic field value is neither MoveObject nor MoveValue")
			}

			if len(balanceJSON) == 0 || string(balanceJSON) == "null" {
				return nil, fmt.Errorf("balance JSON is empty or null for coin type %s", cointype)
			}

			var coinBalance struct {
				Value *iotagraphql.BigInt `json:"value"`
			}
			if err := json.Unmarshal(balanceJSON, &coinBalance); err != nil {
				return nil, fmt.Errorf("failed to unmarshal balance JSON: %w", err)
			}

			bag.SetCoin(cointype, iotagraphql.CoinValue(coinBalance.Value.Uint64()))
		} else {
			// non-coin asset (object, NFT, etc.)
			moveObj, ok := node.Value.(*graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject)
			if !ok {
				return nil, fmt.Errorf("non-coin dynamic field is not a MoveObject")
			}
			typ, err := iotago.ObjectTypeFromString(moveObj.Contents.Type.Repr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ObjectType: %w", err)
			}
			objectID := iotago.ObjectID(moveObj.Address)
			bag.AddObject(objectID, typ)
		}
	}

	return &bag, nil
}
