package cryptolib

import (
	"encoding/json"
)

// jsonAddress defines the json representation of any Address.
type jsonAddress struct {
	Type       int    `json:"type"`
	PubKeyHash string `json:"pubKeyHash"`
}

func AddressToJSON(addr Address) (*json.RawMessage, error) {
	jsonAddress := &jsonAddress{
		Type:       int(addr.Type()),
		PubKeyHash: EncodeHex(addr.Bytes()),
	}
	jsonAddressBytes, err := json.Marshal(jsonAddress)
	if err != nil {
		return nil, err
	}
	jsonAddressRaw := json.RawMessage(jsonAddressBytes)
	return &jsonAddressRaw, nil
}

func AddressFromJSON(bytes *json.RawMessage) (Address, error) {
	jsonAddress := &jsonAddress{}
	if err := json.Unmarshal(*bytes, jsonAddress); err != nil {
		return nil, err
	}
	fullBytes := []byte{byte(jsonAddress.Type)}
	fullBytes = append(fullBytes, jsonAddress.PubKeyHash...)
	return DeserializeAddress(fullBytes)
}
