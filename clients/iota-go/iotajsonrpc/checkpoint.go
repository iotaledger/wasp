package iotajsonrpc

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

type Checkpoint struct {
	Epoch                      *BigInt                       `json:"epoch"`
	SequenceNumber             *BigInt                       `json:"sequenceNumber"`
	Digest                     iotago.Digest                 `json:"digest"`
	NetworkTotalTransactions   *BigInt                       `json:"networkTotalTransactions"`
	PreviousDigest             *iotago.Digest                `json:"previousDigest,omitempty"`
	EpochRollingGasCostSummary GasCostSummary                `json:"epochRollingGasCostSummary"`
	TimestampMs                *BigInt                       `json:"timestampMs"`
	Transactions               []*iotago.Digest              `json:"transactions"`
	CheckpointCommitments      []iotago.CheckpointCommitment `json:"checkpointCommitments"`
	ValidatorSignature         iotago.Base64Data             `json:"validatorSignature"`
}

type CheckpointPage = Page[*Checkpoint, BigInt]
