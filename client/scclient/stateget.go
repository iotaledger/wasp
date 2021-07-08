package scclient

// StateGet fetches the raw value associated with the given key in the smart contract state
func (c *SCClient) StateGet(key string) ([]byte, error) {
	return c.ChainClient.StateGet(string(c.ContractHname.Bytes()) + key)
}
