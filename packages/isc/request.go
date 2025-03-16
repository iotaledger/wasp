package isc

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
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
	Allowance() *Assets // transfer of assets to the smart contract. Debited from sender's L2 account
	Assets() *Assets    // attached assets for the on-ledger request, nil for off-ledger. All goes to sender.
	Message() Message
	GasBudget() (gas uint64, isEVM bool)
	ID() RequestID
	SenderAccount() AgentID
	TargetAddress() *cryptolib.Address
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
