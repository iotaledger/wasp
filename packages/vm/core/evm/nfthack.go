package evm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
)

// This hack is so that the ERC721 tokenURI view function returns the NFT name and description
// for explorers
type PackedNFTURI struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Attributes  string `json:"attributes,omitempty"`
	Image       string `json:"image"`
}

const dataURLPrefix = "data:application/json;base64"

func EncodePackedNFTURI(metadata *isc.IRC27NFTMetadata) string {
	return dataURLPrefix + "," + base64.StdEncoding.EncodeToString(lo.Must(json.Marshal(PackedNFTURI{
		Name:        metadata.Name,
		Description: metadata.Description,
		Image:       metadata.URI,
		Attributes:  metadata.Attributes,
	})))
}

func DecodePackedNFTURI(uri string) (*PackedNFTURI, error) {
	parts := strings.Split(uri, ",")
	if len(parts) != 2 {
		return nil, errors.New("cannot decode packed NFT URI: expected valid data URL")
	}
	if parts[0] != dataURLPrefix {
		return nil, errors.New("cannot decode packed NFT URI: expected valid data URL")
	}
	b, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode packed NFT URI: %w", err)
	}
	var p *PackedNFTURI
	err = json.Unmarshal(b, &p)
	return p, err
}
