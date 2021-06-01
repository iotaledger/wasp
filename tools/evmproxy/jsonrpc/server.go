package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/hive.go/crypto/ed25519"
)

func NewServer(chain *EVMChain, signer *ed25519.KeyPair) *rpc.Server {
	rpcsrv := rpc.NewServer()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"web3", NewWeb3Service()},
		{"net", NewNetService()},
		{"eth", NewEthService(chain, signer)},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			panic(err)
		}
	}
	return rpcsrv
}
