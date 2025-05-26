package models

type OffLedgerRequest struct {
	Request string `json:"request" swagger:"desc(Offledger Request (Hex)),required"`
}

type ContractCallViewRequest struct {
	ContractName  string   `json:"contractName" swagger:"desc(The contract name),required"`
	ContractHName string   `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionName  string   `json:"functionName" swagger:"desc(The function name),required"`
	FunctionHName string   `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
	Arguments     []string `json:"arguments" swagger:"desc(Encoded arguments to be passed to the function),required"`
	Block         string   `json:"block" swagger:"desc(block index or trie root to execute the view call in, latest block will be used if not specified)"`
}

type EstimateGasRequestOnledger struct {
	TransactionBytes string `json:"transactionBytes" swagger:"desc(The result of the DryRun encoded in BCS format and represented in hexadecimal),required"`
}

type EstimateGasRequestOffledger struct {
	Request     string `json:"requestBytes" swagger:"desc(Offledger Request (Hex)),required"`
	FromAddress string `json:"fromAddress" swagger:"desc(The address to estimate gas for(Hex)),required"`
}
