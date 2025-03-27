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
		Assets: iscmove.Assets{
			Coins:   make(iscmove.CoinBalances),
			Objects: make(iscmove.ObjectCollection),
		},
	}
	for _, data := range fields.Data {
		panic("TODO: handle non-coin objects")
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

		var moveBalance iotajsonrpc.MoveBalance
		err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &moveBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields in Balance: %w", err)
		}

		cointype, err := iotajsonrpc.CoinTypeFromString("0x" + data.Name.Value.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to convert cointype from iotajsonrpc: %w", err)
		}

		bag.Coins[cointype] = iotajsonrpc.CoinValue(moveBalance.Value.Uint64())
	}

	return &bag, nil
}
