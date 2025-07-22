package iotajsonrpc

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type CommitteeInfo struct {
	EpochId    *BigInt     `json:"epoch"`
	Validators []Validator `json:"validators"`
}

type Validator struct {
	PublicKey *iotago.Base64Data
	Stake     *BigInt
}

func (c *CommitteeInfo) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var epochSafeBigInt BigInt
	if epochRaw, ok := raw["epoch"].(string); ok {
		if err := epochSafeBigInt.UnmarshalText([]byte(epochRaw)); err != nil {
			return err
		}
		c.EpochId = &epochSafeBigInt
	}

	if validators, ok := raw["validators"].([]interface{}); ok {
		for _, validator := range validators {
			var epochSafeBigInt BigInt
			if validatorElts, ok := validator.([]interface{}); ok && len(validatorElts) == 2 {
				publicKey, err := iotago.NewBase64Data(validatorElts[0].(string))
				if err != nil {
					return err
				}
				if err := epochSafeBigInt.UnmarshalText([]byte(validatorElts[1].(string))); err != nil {
					return err
				}
				c.Validators = append(c.Validators, Validator{
					PublicKey: publicKey,
					Stake:     &epochSafeBigInt,
				})
			}
		}
	}

	return nil
}
