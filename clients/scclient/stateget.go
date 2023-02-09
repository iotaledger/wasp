package scclient

import "context"

// StateGet fetches the raw value associated with the given key in the smart contract state
func (c *SCClient) StateGet(ctx context.Context, key string) ([]byte, error) {
	return c.ChainClient.StateGet(ctx, string(c.ContractHname.Bytes())+key)
}
