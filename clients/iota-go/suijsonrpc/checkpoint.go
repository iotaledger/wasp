package suijsonrpc

import "github.com/iotaledger/wasp/clients/iota-go/sui"

type Checkpoint struct {
	Epoch                      *BigInt                    `json:"epoch"`
	SequenceNumber             *BigInt                    `json:"sequenceNumber"`
	Digest                     sui.Digest                 `json:"digest"`
	NetworkTotalTransactions   *BigInt                    `json:"networkTotalTransactions"`
	PreviousDigest             *sui.Digest                `json:"previousDigest,omitempty"`
	EpochRollingGasCostSummary GasCostSummary             `json:"epochRollingGasCostSummary"`
	TimestampMs                *BigInt                    `json:"timestampMs"`
	Transactions               []*sui.Digest              `json:"transactions"`
	CheckpointCommitments      []sui.CheckpointCommitment `json:"checkpointCommitments"`
	ValidatorSignature         sui.Base64Data             `json:"validatorSignature"`
}

type CheckpointPage = Page[*Checkpoint, BigInt]
