package parameters

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// RentStructure defines the parameters of rent cost calculations on objects which take node resources.
type RentStructure struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor
	// defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor
}

func (r *RentStructure) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	var factorData uint8
	var factorKey uint8

	return serializer.NewDeserializer(data).
		ReadNum(&r.VByteCost, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte cost within rent structure", err)
		}).
		ReadNum(&factorData, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte factor data within rent structure", err)
		}).
		Do(func() {
			r.VBFactorData = VByteCostFactor(factorData)
		}).
		ReadNum(&factorKey, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte factor key within rent structure", err)
		}).
		Do(func() {
			r.VBFactorKey = VByteCostFactor(factorKey)
		}).
		Done()
}

func (r *RentStructure) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(&r.VByteCost, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte cost within rent structure", err)
		}).
		WriteNum(&r.VBFactorData, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte factor data within rent structure", err)
		}).
		WriteNum(&r.VBFactorKey, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte factor key within rent structure", err)
		}).
		Serialize()
}

func (r *RentStructure) MarshalJSON() ([]byte, error) {
	jRentStructure := &jsonRentStructure{
		VByteCost:    r.VByteCost,
		VBFactorData: uint8(r.VBFactorData),
		VBFactorKey:  uint8(r.VBFactorKey),
	}
	return json.Marshal(jRentStructure)
}

func (r *RentStructure) UnmarshalJSON(data []byte) error {
	jRentStructure := &jsonRentStructure{}
	if err := json.Unmarshal(data, jRentStructure); err != nil {
		return err
	}
	seri, err := jRentStructure.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*RentStructure)
	return nil
}

// jsonRentStructure defines the json representation of a RentStructure.
type jsonRentStructure struct {
	VByteCost    uint32 `json:"vByteCost"`
	VBFactorData uint8  `json:"vByteFactorData"`
	VBFactorKey  uint8  `json:"vByteFactorKey"`
}

func (j *jsonRentStructure) ToSerializable() (serializer.Serializable, error) {
	return &RentStructure{
		VByteCost:    j.VByteCost,
		VBFactorData: VByteCostFactor(j.VBFactorData),
		VBFactorKey:  VByteCostFactor(j.VBFactorKey),
	}, nil
}

// CoversStateRent tells whether given this NonEphemeralObject, the given rent fulfills the renting costs
// by examining the virtual bytes cost of the object.
// Returns the minimum rent computed and an error if it is not covered by rent.
/*func (r *RentStructure) CoversStateRent(object NonEphemeralObject, rent uint64) (uint64, error) {
	minRent := r.MinRent(object)
	if rent < minRent {
		return 0, fmt.Errorf("%w: needed %d but only got %d", ErrVByteRentNotCovered, minRent, rent)
	}
	return minRent, nil
}*/

// MinRent returns the minimum rent to cover a given object.
/*func (r *RentStructure) MinRent(object NonEphemeralObject) uint64 {
	return uint64(r.VByteCost) * uint64(object.VBytes(r, nil))
}*/

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
/*func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) uint64 {
	return uint64(r.VByteCost) * uint64((&BasicOutput{Conditions: UnlockConditions{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil))
}*/
