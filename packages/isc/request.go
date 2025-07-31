package isc

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

// Request wraps any data which can be potentially be interpreted as a request
type Request interface {
	Calldata

	Bytes() []byte
	IsOffLedger() bool
	String() string
}

func init() {
	bcs.RegisterEnumType5[Request, *OnLedgerRequestData, *OffLedgerRequestData, *evmOffLedgerTxRequest, *evmOffLedgerCallRequest, *ImpersonatedOffLedgerRequestData]()
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
	// Assets returns the attached assets for the on-ledger request, empty for off-ledger.
	// Attached assets are deposited in the sender's L2 account by default.
	Assets() *Assets
	// Allowance returns the assets that the sender allows to be debited
	// from their L2 account and transferred to the target contract
	// Returns error if there was an error decoding the allowance from the on-ledger request.
	Allowance() (*Assets, error)
	Message() Message
	GasBudget() (gas uint64, isEVM bool)
	ID() RequestID
	SenderAccount() AgentID
	EVMCallMsg() *ethereum.CallMsg
}

type UnsignedOffLedgerRequest interface {
	Bytes() []byte
	WithNonce(nonce uint64) UnsignedOffLedgerRequest
	WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest
	WithAllowance(allowance *Assets) UnsignedOffLedgerRequest
	WithSender(sender *cryptolib.PublicKey) OffLedgerRequest
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

// RequestGasPrice returns:
// for ISC request: nil,
// for EVM tx: the gas price set in the EVM tx (full decimals), or 0 if gas price is unset
func RequestGasPrice(req Request) *big.Int {
	callMsg := req.EVMCallMsg()
	if callMsg == nil {
		return nil
	}
	if callMsg.GasPrice == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(callMsg.GasPrice)
}
