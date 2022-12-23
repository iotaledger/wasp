// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func Client(apiAddress string) *chainclient.Client {
	return chainclient.New(
		config.L1Client(),
		config.WaspClient(apiAddress),
		GetCurrentChainID(),
		wallet.Load().KeyPair,
	)
}

func SCClient(contractHname isc.Hname, i ...int) *scclient.SCClient {
	return scclient.New(Client(config.WaspAPI(i...)), contractHname)
}
