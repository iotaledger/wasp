package testplugins

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/mr-tron/base58"
)

// BLS addresses
const (
	blsaddr1 = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"
	blsaddr2 = "dV9hfYyHq7uiCKdKYQoLqyiwX6tN448GRm8UgFpUC3Vo"
	blsaddr3 = "eiMbhrJjajqnCLmVJqFXzFsh1ZsbCAnJ9wauU8cP8uxL"
)

// owner address
const (
	ownerPrivateKey = "2DC3vbiBF9EE1bUP4eUjrzghxnf8xB1r6GAVKMJM7NsznTcqqTUUYhzkRQe2vae7YViiACcUFM8bVUEdiyp59tFY"
	ownerPublicKey  = "AM2n1SpSEUn68spvecrr14MrqsK1jp6rzfy2myHUXsLa"
	ownerAddressStr = "XnJ1a4V8P7sEaDp7sRnpnisP8uF9ZDMQ38sp5hzUUAA5"
	dummyTxIdStr    = "C1q8mjxNpMJJ8E6xicxEuDVU5FvsWfqjeLFxz561eDx8"

	color1str = "CGWjDfPaik2Ahgzs4oiiywmWHtLV5jVadMiUZGvkLXRQ"
	color2str = "5bHVubzwPWadUgFa3xA7vtWtr7disVHCYc5nYChTqz9j"
	color3str = "DrSQHAg8fkCCfsCE2xSkygVTsGzcXC8a7J4yEJbnaizJ"
)

var (
	ownerAddress   address.Address
	ownerKeyPair   ed25519.KeyPair
	ownerSigScheme signaturescheme.SignatureScheme
	addr1          address.Address
	addr2          address.Address
	addr3          address.Address
	SC1            apilib.NewOriginParams
	SC2            apilib.NewOriginParams
	SC3            apilib.NewOriginParams
	dummyTxId      valuetransaction.ID
)

func init() {
	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	var err error
	priv, err := base58.Decode(ownerPrivateKey)
	if err != nil || len(priv) != len(privKey) {
		panic(err)
	}
	pub, err := base58.Decode(ownerPublicKey)
	if err != nil || len(pub) != len(pubKey) {
		panic(err)
	}
	copy(privKey[:], priv)
	copy(pubKey[:], pub)
	ownerKeyPair = ed25519.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
	ownerSigScheme = signaturescheme.ED25519(ownerKeyPair)
	ownerAddress, err = address.FromBase58(ownerAddressStr)
	if err != nil {
		panic(err)
	}
	if ownerSigScheme.Address() != ownerAddress {
		panic("ownerSigScheme.Address() != ownerAddress")
	}

	addr1, err = address.FromBase58(blsaddr1)
	if err != nil {
		panic(err)
	}
	addr2, err = address.FromBase58(blsaddr2)
	if err != nil {
		panic(err)
	}
	addr3, err = address.FromBase58(blsaddr3)
	if err != nil {
		panic(err)
	}
	dummyTxId, err = sctransaction.TransactionIDFromString(dummyTxIdStr)
	if err != nil {
		panic(err)
	}
	SC1 = apilib.NewOriginParams{
		Address:      &addr1,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 1 one",
	}
	SC1.ProgramHash = hashing.HashStrings(SC1.Description)

	SC2 = apilib.NewOriginParams{
		Address:      &addr2,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 2 two",
	}
	SC2.ProgramHash = hashing.HashStrings(SC2.Description)

	SC3 = apilib.NewOriginParams{
		Address:      &addr3,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 3 three",
	}
	SC3.ProgramHash = hashing.HashStrings(SC3.Description)

	col1, err := sctransaction.ColorFromString(color1str)
	if err != nil {
		panic(err)
	}
	col2, err := sctransaction.ColorFromString(color2str)
	if err != nil {
		panic(err)
	}
	col3, err := sctransaction.ColorFromString(color3str)
	if err != nil {
		panic(err)
	}

	otx1, _ := CreateOriginData(SC1, nil)
	otx2, _ := CreateOriginData(SC2, nil)
	otx3, _ := CreateOriginData(SC3, nil)

	if !(otx1.ID() == valuetransaction.ID(col1)) {
		panic("assertion failed: otx1.ID() == valuetransaction.ID(col1)")
	}
	if !(otx2.ID() == valuetransaction.ID(col2)) {
		panic("assertion failed: otx2.ID() == valuetransaction.ID(col2)")
	}
	if !(otx3.ID() == valuetransaction.ID(col3)) {
		panic("assertion failed: otx3.ID() == valuetransaction.ID(col3)")
	}
}

func CreateOriginData(par apilib.NewOriginParams, nodeLocations []string) (*valuetransaction.Transaction, *registry.SCMetaData) {
	tx := apilib.NewOriginTransaction(par, &dummyTxId, ownerSigScheme)
	if nodeLocations == nil {
		return tx, nil
	}
	scdata := &registry.SCMetaData{
		Address:       *par.Address,
		Color:         balance.Color(tx.ID()),
		OwnerAddress:  *par.OwnerAddress,
		Description:   par.Description,
		ProgramHash:   *par.ProgramHash,
		NodeLocations: nodeLocations,
	}
	return tx, scdata
}
