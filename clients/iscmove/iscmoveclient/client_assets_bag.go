package iscmoveclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
)

func (c *Client) GetAssetsBagWithBalances(
	ctx context.Context,
	assetsBagID *iotago.ObjectID,
) (*iscmove.AssetsBagWithBalances, error) {
	fields, err := c.GetDynamicFields(ctx, iotagraphql.GetDynamicFieldsRequest{ParentObjectID: assetsBagID})
	if err != nil {
		return nil, fmt.Errorf("failed to get DynamicFields in AssetsBag: %w", err)
	}

	bag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *assetsBagID,
			Size: uint64(len(fields.Data)),
		},
		Assets: *iscmove.NewEmptyAssets(),
	}
	for _, data := range fields.Data {
		// for coins the "field name" is of type 0x1::ascii::String
		// for non-coins it's 0x2::object::ID
		isCoin, err := iotago.IsSameResource(data.Name.Type, "0x1::ascii::String")
		if err != nil {
			return nil, fmt.Errorf("failed to check if resource is coin: %w", err)
		}

		if isCoin {
			// Convert coin type from the dynamic field name
			cointype, err := iotagraphql.CoinTypeFromString("0x" + data.Name.Value.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to convert cointype: %w", err)
			}

			var balanceJSON []byte

			// Check if it's a DynamicObject or DynamicField
			if data.Type.Data.DynamicObject != nil {
				// DynamicObject: use GetObject with the ObjectID
				resGetObject, err2 := c.GetObject(ctx, iotagraphql.GetObjectRequest{
					ObjectID: &data.ObjectID,
					Options:  &iotagraphql.IotaObjectDataOptions{ShowContent: true},
				})
				if err2 != nil {
					return nil, fmt.Errorf("failed to call GetObject for Balance (coin type %s): %w", cointype, err2)
				}

				if resGetObject.Data == nil || resGetObject.Data.Content == nil || resGetObject.Data.Content.Data.MoveObject == nil {
					return nil, fmt.Errorf("content data of AssetBag nil! (%s)", assetsBagID)
				}

				balanceJSON = resGetObject.Data.Content.Data.MoveObject.Fields
			} else if data.Type.Data.DynamicField != nil {
				// DynamicField: extract the value directly from the ValueJson field
				if len(data.ValueJson) == 0 {
					return nil, fmt.Errorf("ValueJson is empty for wrapped dynamic field (coin type %s)", cointype)
				}
				balanceJSON = data.ValueJson
			} else {
				return nil, fmt.Errorf("coin dynamic field is neither DynamicObject nor DynamicField: %+v", data)
			}

			// Check if balance JSON is empty or null
			if len(balanceJSON) == 0 || string(balanceJSON) == "null" {
				return nil, fmt.Errorf("balance JSON is empty or null for coin type %s", cointype)
			}

			var coinBalance struct {
				Value *iotagraphql.BigInt `json:"value"`
			}

			err = json.Unmarshal(balanceJSON, &coinBalance)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal balance JSON: %w", err)
			}

			bag.SetCoin(cointype, iotagraphql.CoinValue(coinBalance.Value.Uint64()))
		} else {
			// non-coin asset (i.e. an "object", nft, etc)
			typ, err := iotago.ObjectTypeFromString(data.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ObjectType: %w", err)
			}
			bag.AddObject(data.ObjectID, typ)
		}
	}

	return &bag, nil
}
