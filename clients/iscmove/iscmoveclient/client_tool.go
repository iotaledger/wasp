package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

func (c *Client) GetCoin(
	ctx context.Context,
	coinID *iotago.ObjectID,
) (*MoveCoin, error) {
	getCoinRes, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{
		ObjectID: coinID,
		Options:  &iotagraphql.IotaObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject: %w", err)
	}
	var moveCoin MoveCoin
	err = iotagraphql.UnmarshalBCS(getCoinRes.Data.Bcs.MoveObject.BcsBytes, &moveCoin)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhal MoveCoin: %w", err)
	}
	return &moveCoin, nil
}
