// +build ignore

package chainclient

import "github.com/iotaledger/wasp/packages/webapi/model/statequery"

// StateQuery queries the chain state, and returns the result of the query.
func (c *Client) StateQuery(query *statequery.Request) (*statequery.Results, error) {
	return c.WaspClient.StateQuery(&c.ChainID, query)
}
