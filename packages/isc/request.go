package isc

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// Request wraps any data which can be potentially be interpreted as a request
type Request interface {
	Calldata

	Bytes() []byte
	IsOffLedger() bool
	String() string

	Read(r io.Reader) error
	Write(w io.Writer) error
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
	Allowance() *Assets // transfer of assets to the smart contract. Debited from sender account
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
	Clone() OnLedgerRequest
	RequestRef() sui.ObjectRef
	AssetsBag() *iscmove.AssetsBag
}

func MustLogRequestsInTransaction(tx *iotago.Transaction, log func(msg string, args ...interface{}), prefix string) {
	txReqs, err := RequestsInTransaction(tx)
	if err != nil {
		panic(fmt.Errorf("cannot extract requests from TX: %w", err))
	}
	for chainID, chainReqs := range txReqs {
		for i, req := range chainReqs {
			log("%v, ChainID=%v, Req[%v]=%v", prefix, chainID.ShortString(), i, req.String())
		}
	}
}

// TODO: Refactor me:
// RequestsInTransaction parses the transaction and extracts those outputs which are interpreted as a request to a chain
func RequestsInTransaction(tx *iotago.Transaction) (map[ChainID][]Request, error) {
	txid, err := tx.ID()
	if err != nil {
		return nil, err
	}
	if tx.Essence == nil {
		return nil, fmt.Errorf("malformed transaction")
	}

	ret := make(map[ChainID][]Request)
	_ = txid
	panic("refactor me")
	/*for i, output := range tx.Essence.Outputs {
		switch output.(type) {
		case *iotago.BasicOutput, *iotago.NFTOutput:
			// process it
		default:
			// only BasicOutputs and NFTs are interpreted right now, // TODO other outputs
			continue
		}

		// wrap request into the isc.Request
		odata, err := OnLedgerFromUTXO(output, iotago.OutputIDFromTransactionIDAndIndex(txid, uint16(i)))
		if err != nil {
			return nil, err // TODO: maybe log the error and keep processing?
		}

		addr := odata.TargetAddress()


		chainID := ChainIDFromAddress(addr)

		if odata.IsInternalUTXO(chainID) {
			continue
		}

		ret[chainID] = append(ret[chainID], odata)
	}*/
	return ret, nil
}

// TODO: Clarify if we want to keep expiry dates.
// don't process any request which deadline will expire within 1 minute
/*const RequestConsideredExpiredWindow = time.Minute * 1
func RequestIsExpired(req OnLedgerRequest, currentTime time.Time) bool {
	expiry, _ := req.Features().Expiry()
	if expiry.IsZero() {
		return false
	}
	return !expiry.IsZero() && currentTime.After(expiry.Add(-RequestConsideredExpiredWindow))
}*/

func RequestHash(req Request) hashing.HashValue {
	return hashing.HashData(req.Bytes())
}
