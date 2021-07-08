package chainclient

// StateGet fetches the raw value associated with the given key in the chain state
func (c *Client) StateGet(key string) ([]byte, error) {
	return c.WaspClient.StateGet(c.ChainID, key)
}
