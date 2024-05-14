package parameters

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

var (
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = errors.New("serializable is not a supported rent structure")
)

var (
	protocolParametersRentStructureGuard = &serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			return &RentStructure{}, nil
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedRentStructure)
			}

			return nil
		},
	}
)

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte
	// The human friendly name of the network.
	NetworkName string
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix
	// The minimum pow score of the network.
	MinPoWScore uint32
	// The below max depth parameter of the network.
	BelowMaxDepth uint8
	// The rent structure used by given node/network.
	RentStructure RentStructure
	// TokenSupply defines the current token supply on the network.
	TokenSupply uint64
}

func (p *ProtocolParameters) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	var bech32HRP string
	var rentStructure *RentStructure
	return serializer.NewDeserializer(data).
		ReadByte(&p.Version, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize version within protocol parameters", err)
		}).
		ReadString(&p.NetworkName, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize network name within protocol parameters", err)
		}, 0, 0).
		ReadString(&bech32HRP, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize Bech32HRP prefix within protocol parameters", err)
		}, 0, 0).
		Do(func() {
			p.Bech32HRP = NetworkPrefix(bech32HRP)
		}).
		ReadNum(&p.MinPoWScore, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize minimum pow score within protocol parameters", err)
		}).
		ReadNum(&p.BelowMaxDepth, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize below max depth within protocol parameters", err)
		}).
		ReadObject(&rentStructure, deSeriMode, deSeriCtx, serializer.TypeDenotationNone, protocolParametersRentStructureGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize rent structure within protocol parameters", err)
		}).
		Do(func() {
			p.RentStructure = *rentStructure
		}).
		ReadNum(&p.TokenSupply, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize token supply within protocol parameters", err)
		}).
		Done()
}

func (p *ProtocolParameters) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteByte(p.Version, func(err error) error {
			return fmt.Errorf("%w: unable to serialize version within protocol parameters", err)
		}).
		WriteString(p.NetworkName, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to serialize network name within protocol parameters", err)
		}, 0, 0).
		WriteString(string(p.Bech32HRP), serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to serialize Bech32HRP prefix within protocol parameters", err)
		}, 0, 0).
		WriteNum(p.MinPoWScore, func(err error) error {
			return fmt.Errorf("%w: unable to serialize minimum pow score within protocol parameters", err)
		}).
		WriteNum(p.BelowMaxDepth, func(err error) error {
			return fmt.Errorf("%w: unable to serialize below max depth within protocol parameters", err)
		}).
		WriteObject(&p.RentStructure, deSeriMode, deSeriCtx, protocolParametersRentStructureGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("%w: unable to serialize rent structure within protocol parameters", err)
		}).
		WriteNum(p.TokenSupply, func(err error) error {
			return fmt.Errorf("%w: unable to serialize token supply within protocol parameters", err)
		}).
		Serialize()
}

func (p ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.NetworkName)
}
