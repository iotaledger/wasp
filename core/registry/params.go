package registry

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersRegistries struct {
	Chains struct {
		FilePath string `default:"waspdb/chains/chain_registry.json" usage:"the path to the chain registry file"`
	}
	DKShares struct {
		Path string `default:"waspdb/dkshares" usage:"the path to the distributed key shares registries folder"`
	}
	TrustedPeers struct {
		FilePath string `default:"waspdb/trusted_peers.json" usage:"the path to the trusted peers registry file"`
	}
	ConsensusState struct {
		Path string `default:"waspdb/chains/consensus" usage:"the path to the consensus state registries folder"`
	}
}

// ParametersP2P contains the definition of the parameters used by p2p.
type ParametersP2P struct {
	Identity struct {
		PrivateKey string `default:"" usage:"private key used to derive the node identity (optional)"`
		FilePath   string `default:"waspdb/identity/identity.key" usage:"the path to the node identity PEM file"`
	}

	Database struct {
		// Defines the path to the p2p database.
		Path string `default:"waspdb/p2pstore" usage:"the path to the p2p database"`
	} `name:"db"`
}

var (
	ParamsRegistries = &ParametersRegistries{}
	ParamsP2P        = &ParametersP2P{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"registries": ParamsRegistries,
		"p2p":        ParamsP2P,
	},
	Masked: []string{"p2p.identity.privateKey"},
}
