package iscmoveclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
)

func (c *Client) GetAssetsBagWithBalances(
	ctx context.Context,
	assetsBagID *iotago.ObjectID,
) (*iscmove.AssetsBagWithBalances, error) {
	fields, err := c.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{ParentObjectID: assetsBagID})
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
			resGetObject, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
				ObjectID: &data.ObjectID,
				Options:  &iotajsonrpc.IotaObjectDataOptions{ShowContent: true},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to call GetObject for Balance: %w", err)
			}

			if resGetObject.Data.Content == nil || resGetObject.Data.Content.Data.MoveObject == nil {
				return nil, fmt.Errorf("content data of AssetBag nil! (%s)", assetsBagID)
			}
			var coinBalance struct {
				ID    *iotajsonrpc.MoveUID
				Name  *iotago.ResourceType
				Value *iotajsonrpc.BigInt
			}

			err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &coinBalance)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
			}

			cointype, err := iotajsonrpc.CoinTypeFromString("0x" + data.Name.Value.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to convert cointype from iotajsonrpc: %w", err)
			}

			bag.Coins[cointype] = iotajsonrpc.CoinValue(coinBalance.Value.Uint64())
		} else {
			// non-coin asset (i.e. an "object", nft, etc)
			typ, err := iotago.ObjectTypeFromString(data.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ObjectType: %w", err)
			}
			bag.Objects[data.ObjectID] = typ
		}
	}

	return &bag, nil
}
