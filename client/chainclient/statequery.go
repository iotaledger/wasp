package chainclient

import "github.com/iotaledger/wasp/packages/webapi/model/statequery"

func (c *Client) StateQuery(query *statequery.Request) (*statequery.Results, error) {
	return c.WaspClient.StateQuery(&c.ChainID, query)
}
