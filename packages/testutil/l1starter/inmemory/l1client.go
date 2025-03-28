package inmemory

import (
	`context`
	`time`

	`github.com/iotaledger/wasp/clients`
	`github.com/iotaledger/wasp/clients/iota-go/iotaclient`
	`github.com/iotaledger/wasp/clients/iota-go/iotago`
	`github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc`
	`github.com/iotaledger/wasp/clients/iota-go/iotasigner`
	`github.com/iotaledger/wasp/packages/cryptolib`
	iotasimulator `github.com/lmoe/iota-simulator`
)

type InMemoryL1Client struct {
	sim *iotasimulator.Simulator
}

func (i InMemoryL1Client) GetDynamicFieldObject(ctx context.Context, req iotaclient.GetDynamicFieldObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetDynamicFields(ctx context.Context, req iotaclient.GetDynamicFieldsRequest) (*iotajsonrpc.DynamicFieldPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetOwnedObjects(ctx context.Context, req iotaclient.GetOwnedObjectsRequest) (*iotajsonrpc.ObjectsPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) QueryEvents(ctx context.Context, req iotaclient.QueryEventsRequest) (*iotajsonrpc.EventPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) QueryTransactionBlocks(ctx context.Context, req iotaclient.QueryTransactionBlocksRequest) (*iotajsonrpc.TransactionBlocksPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) ResolveNameServiceNames(ctx context.Context, req iotaclient.ResolveNameServiceNamesRequest) (*iotajsonrpc.IotaNamePage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) DevInspectTransactionBlock(ctx context.Context, req iotaclient.DevInspectTransactionBlockRequest) (*iotajsonrpc.DevInspectResults, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) DryRunTransaction(ctx context.Context, txDataBytes iotago.Base64Data) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) ExecuteTransactionBlock(ctx context.Context, req iotaclient.ExecuteTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCommitteeInfo(ctx context.Context, epoch *iotajsonrpc.BigInt) (*iotajsonrpc.CommitteeInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) BatchTransaction(ctx context.Context, req iotaclient.BatchTransactionRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MergeCoins(ctx context.Context, req iotaclient.MergeCoinsRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MoveCall(ctx context.Context, req iotaclient.MoveCallRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) Pay(ctx context.Context, req iotaclient.PayRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) PayAllIota(ctx context.Context, req iotaclient.PayAllIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) PayIota(ctx context.Context, req iotaclient.PayIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) Publish(ctx context.Context, req iotaclient.PublishRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) RequestAddStake(ctx context.Context, req iotaclient.RequestAddStakeRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) RequestWithdrawStake(ctx context.Context, req iotaclient.RequestWithdrawStakeRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) SplitCoin(ctx context.Context, req iotaclient.SplitCoinRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) SplitCoinEqual(ctx context.Context, req iotaclient.SplitCoinEqualRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) TransferObject(ctx context.Context, req iotaclient.TransferObjectRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) TransferIota(ctx context.Context, req iotaclient.TransferIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCoinObjsForTargetAmount(ctx context.Context, address *iotago.Address, targetAmount uint64, gasAmount uint64) (iotajsonrpc.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) SignAndExecuteTransaction(ctx context.Context, req *iotaclient.SignAndExecuteTransactionRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) PublishContract(ctx context.Context, signer iotasigner.Signer, modules []*iotago.Base64Data, dependencies []*iotago.Address, gasBudget uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) UpdateObjectRef(ctx context.Context, ref *iotago.ObjectRef) (*iotago.ObjectRef, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MintToken(ctx context.Context, signer iotasigner.Signer, packageID *iotago.PackageID, tokenName string, treasuryCap *iotago.ObjectRef, mintAmount uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) BatchGetObjectsOwnedByAddress(ctx context.Context, address *iotago.Address, options *iotajsonrpc.IotaObjectDataOptions, filterType string) ([]iotajsonrpc.IotaObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) BatchGetFilteredObjectsOwnedByAddress(ctx context.Context, address *iotago.Address, options *iotajsonrpc.IotaObjectDataOptions, filter func(*iotajsonrpc.IotaObjectData) bool) ([]iotajsonrpc.IotaObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetTotalSupply(ctx context.Context, coinType iotago.ObjectType) (*iotajsonrpc.Supply, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetChainIdentifier(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCheckpoint(ctx context.Context, checkpointId *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetProtocolConfig(ctx context.Context, version *iotajsonrpc.BigInt) (*iotajsonrpc.ProtocolConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.IotaObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MultiGetTransactionBlocks(ctx context.Context, req iotaclient.MultiGetTransactionBlocksRequest) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) TryGetPastObject(ctx context.Context, req iotaclient.TryGetPastObjectRequest) (*iotajsonrpc.IotaPastObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) TryMultiGetPastObjects(ctx context.Context, req iotaclient.TryMultiGetPastObjectsRequest) ([]*iotajsonrpc.IotaPastObjectResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) Health(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) L2() clients.L2Client {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) IotaClient() *iotaclient.Client {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) FindCoinsForGasPayment(ctx context.Context, owner *iotago.Address, pt iotago.ProgrammableTransaction, gasPrice uint64, gasBudget uint64) ([]*iotago.ObjectRef, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) MergeCoinsAndExecute(ctx context.Context, owner iotasigner.Signer, destinationCoin *iotago.ObjectRef, sourceCoins []*iotago.ObjectRef, gasBudget uint64) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) SignAndExecuteTxWithRetry(ctx context.Context, signer iotasigner.Signer, pt iotago.ProgrammableTransaction, gasCoin *iotago.ObjectRef, gasBudget uint64, gasPrice uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i InMemoryL1Client) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	//TODO implement me
	panic("implement me")
}

func NewInMemoryClient(sim *iotasimulator.Simulator) *InMemoryL1Client {
	return &InMemoryL1Client{
		sim: sim,
	}
}
