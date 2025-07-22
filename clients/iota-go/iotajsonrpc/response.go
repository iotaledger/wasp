package iotajsonrpc

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
)

type AuthSignInfo interface{}

type CertifiedTransaction struct {
	TransactionDigest string        `json:"transactionDigest"`
	TxSignature       string        `json:"txSignature"`
	AuthSignInfo      *AuthSignInfo `json:"authSignInfo"`

	Data *SenderSignedData `json:"data"`
}

type ParsedTransactionResponse interface{}

type ExecuteTransactionEffects struct {
	TransactionEffectsDigest string `json:"transactionEffectsDigest"`

	Effects      serialization.TagJson[IotaTransactionBlockEffects] `json:"effects"`
	AuthSignInfo *AuthSignInfo                                      `json:"authSignInfo"`
}

type ExecuteTransactionResponse struct {
	Certificate CertifiedTransaction      `json:"certificate"`
	Effects     ExecuteTransactionEffects `json:"effects"`

	ConfirmedLocalExecution bool `json:"confirmed_local_execution"`
}

func (r *ExecuteTransactionResponse) TransactionDigest() string {
	return r.Certificate.TransactionDigest
}

type IotaCoinMetadata struct {
	Name        string           `json:"name"`
	Symbol      string           `json:"symbol"`
	Decimals    uint8            `json:"decimals"`
	Description string           `json:"description"`
	IconUrl     string           `json:"iconUrl,omitempty"`
	Id          *iotago.ObjectID `json:"id"`
}
