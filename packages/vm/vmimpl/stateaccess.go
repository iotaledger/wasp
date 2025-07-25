package vmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/subrealm"
	"github.com/iotaledger/wasp/v2/packages/vm/execution"
)

func (reqctx *requestContext) chainState(gasBurn bool) kv.KVStore {
	if gasBurn {
		return reqctx.chainStateWithGasBurn()
	}
	return reqctx.uncommittedState
}

func (reqctx *requestContext) chainStateWithGasBurn() kv.KVStore {
	return execution.NewKVStoreWithGasBurn(reqctx.uncommittedState, reqctx)
}

func (reqctx *requestContext) contractStateWithGasBurn() kv.KVStore {
	return subrealm.New(reqctx.chainStateWithGasBurn(), kv.Key(reqctx.CurrentContractHname().Bytes()))
}

func (reqctx *requestContext) ContractStateReaderWithGasBurn() kv.KVStoreReader {
	return subrealm.NewReadOnly(reqctx.chainStateWithGasBurn(), kv.Key(reqctx.CurrentContractHname().Bytes()))
}

func (reqctx *requestContext) SchemaVersion() isc.SchemaVersion {
	return reqctx.vm.schemaVersion
}
