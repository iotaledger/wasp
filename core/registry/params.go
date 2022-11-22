package registry

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersRegistry struct {
	Chains struct {
		FilePath string `default:"waspdb/chain_registry.json" usage:"the path to the chain registry file"`
	}
	DKShares struct {
		FilePath string `default:"waspdb/dkshares.json" usage:"the path to the distributed key shares registry file"`
	}
	TrustedPeers struct {
		FilePath string `default:"waspdb/trusted_peers.json" usage:"the path to the trusted peers registry file"`
	}
}

// ParametersP2P contains the definition of the parameters used by p2p.
type ParametersP2P struct {
	// Defines the private key used to derive the node identity (optional).
	IdentityPrivateKey string `default:"" usage:"private key used to derive the node identity (optional)"`

	Database struct {
		// Defines the path to the p2p database.
		Path string `default:"waspdb/p2pstore" usage:"the path to the p2p database"`
	} `name:"db"`
}

var (
	ParamsRegistry = &ParametersRegistry{}
	ParamsP2P      = &ParametersP2P{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"registry": ParamsRegistry,
		"p2p":      ParamsP2P,
	},
	Masked: []string{"p2p.identityPrivateKey"},
}
