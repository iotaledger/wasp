package nodeconntest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/mr-tron/base58"
)

// BLS addresses
const (
	addr1 = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"
	addr2 = "dV9hfYyHq7uiCKdKYQoLqyiwX6tN448GRm8UgFpUC3Vo"
	addr3 = "eiMbhrJjajqnCLmVJqFXzFsh1ZsbCAnJ9wauU8cP8uxL"
)

// owner address
const (
	ownerPrivateKey = "2DC3vbiBF9EE1bUP4eUjrzghxnf8xB1r6GAVKMJM7NsznTcqqTUUYhzkRQe2vae7YViiACcUFM8bVUEdiyp59tFY"
	ownerPublicKey  = "AM2n1SpSEUn68spvecrr14MrqsK1jp6rzfy2myHUXsLa"
	ownerAddress    = "XnJ1a4V8P7sEaDp7sRnpnisP8uF9ZDMQ38sp5hzUUAA5"
)

func ownerAddressOk() bool {
	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	priv, err := base58.Decode(ownerPrivateKey)
	if err != nil || len(priv) != len(privKey) {
		return false
	}
	pub, err := base58.Decode(ownerPublicKey)
	if err != nil || len(pub) != len(pubKey) {
		return false
	}
	copy(privKey[:], priv)
	copy(pubKey[:], pub)
	keyPair := ed25519.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
	sigScheme := signaturescheme.ED25519(keyPair)
	addr, err := address.FromBase58(ownerAddress)
	if err != nil {
		return false
	}
	return sigScheme.Address() == addr
}

func createOriginTx() *sctransaction.Transaction {
	ownerAddr, _ := address.FromBase58(ownerAddress)
	scAddr, _ := address.FromBase58(addr2)

	txb := sctransaction.NewTransactionBuilder()
	var nilId valuetransaction.ID
	inp1 := valuetransaction.NewOutputID(ownerAddr, nilId)
	txb.AddInputs(inp1)

	txb.AddBalanceToOutput(scAddr, balance.New(balance.ColorNew, 1))

	var col balance.Color = balance.ColorNew
	txb.AddStateBlock(&col, 0)

	h0 := state.OriginVariableStateHash(state.NewOriginBatchParams{
		Address:      &scAddr,
		OwnerAddress: &ownerAddr,
		Description:  "Test SC #2",
		ProgramHash:  hashing.HashStrings("dummy program"),
	})
	txb.SetVariableStateHash(h0)

	ret, err := txb.Finalize()
	if err != nil {
		panic(err)
	}
	return ret
}
