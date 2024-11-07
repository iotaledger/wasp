package isc

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// Request wraps any data which can be potentially be interpreted as a request
type Request interface {
	Calldata

	Bytes() []byte
	IsOffLedger() bool
	String() string
}

func init() {
	bcs.RegisterEnumType4[Request, *OnLedgerRequestData, *OffLedgerRequestData, *evmOffLedgerTxRequest, *evmOffLedgerCallRequest]()
}

func EVMCallDataFromTx(tx *types.Transaction) *ethereum.CallMsg {
	return &ethereum.CallMsg{
		From:       evmutil.MustGetSender(tx),
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: tx.AccessList(),
	}
}

type Calldata interface {
	Allowance() *Assets // transfer of assets to the smart contract. Debited from sender's L2 account
	Assets() *Assets    // attached assets for the on-ledger request, nil for off-ledger. All goes to sender.
	Message() Message
	GasBudget() (gas uint64, isEVM bool)
	ID() RequestID
	SenderAccount() AgentID
	TargetAddress() *cryptolib.Address // TODO implement properly. Target depends on time assumptions and UTXO type
	EVMCallMsg() *ethereum.CallMsg
}

type UnsignedOffLedgerRequest interface {
	Bytes() []byte
	WithNonce(nonce uint64) UnsignedOffLedgerRequest
	WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest
	WithAllowance(allowance *Assets) UnsignedOffLedgerRequest
	WithSender(sender *cryptolib.PublicKey) UnsignedOffLedgerRequest
	Sign(signer cryptolib.Signer) OffLedgerRequest
}

type ImpersonatedOffLedgerRequest interface {
	WithSenderAddress(senderAddress *cryptolib.Address) OffLedgerRequest
}

type OffLedgerRequest interface {
	Request
	ChainID() ChainID
	Nonce() uint64
	VerifySignature() error
	GasPrice() *big.Int
}

type OnLedgerRequest interface {
	Request
	RequestRef() iotago.ObjectRef
	AssetsBag() *iscmove.AssetsBagWithBalances
}

func init() {
	bcs.RegisterEnumType1[OnLedgerRequest, *OnLedgerRequestData]()
}

func RequestHash(req Request) hashing.HashValue {
	return hashing.HashData(req.Bytes())
}
