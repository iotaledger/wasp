// // WaitUntilRequestProcessed blocks until the request has been processed by the node
// func (c *WaspClient) WaitUntilRequestProcessed(chainID *isc.ChainID, reqID isc.RequestID, timeout time.Duration) (*isc.Receipt, error) {
// 	if timeout == 0 {
// 		timeout = reqstatus.WaitRequestProcessedDefaultTimeout
// 	}
// 	var res model.RequestReceiptResponse
// 	err := c.do(
// 		http.MethodGet,
// 		routes.WaitRequestProcessed(chainID.String(), reqID.String()),
// 		&model.WaitRequestProcessedParams{Timeout: timeout},
// 		&res,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var receipt isc.Receipt
// 	err = json.Unmarshal([]byte(res.Receipt), &receipt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &receipt, nil
// }
