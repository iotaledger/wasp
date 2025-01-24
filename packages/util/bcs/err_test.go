package bcs_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/stretchr/testify/require"
)

func TestReadingDefferedErrorHandling(t *testing.T) {
	// FIRST PART:
	// This part just demonstrates buffer works for the second part of test

	var buff bytes.Buffer
	_, err := buff.ReadByte()
	require.Error(t, err) // buffer is empty, so we can expect error

	buff.WriteByte(1)
	_, err = buff.ReadByte()
	require.NoError(t, err) // buffer is not empty, so we expect no error
	_, err = buff.ReadByte()
	require.Error(t, err) // buffer is empty again, so again we expect error
	buff.WriteByte(1)
	_, err = buff.ReadByte()
	require.NoError(t, err) // buffer is not empty again, so again we expect no error

	// SECOND PART:
	buff.WriteByte(1)

	d := bcs.NewDecoder(&buff)
	d.ReadByte()
	require.NoError(t, d.Err())
	d.ReadByte()
	require.Error(t, d.Err()) // buffer is emptry, so error

	buff.WriteByte(1) // making buffer non-empty again
	d.ReadByte()
	require.Error(t, d.Err()) // should be still error once decoder faced error
}

func TestDecodingDefferedErrorHandling(t *testing.T) {
	type TestStruct struct {
		A uint64
		B string
	}

	s := TestStruct{A: 42, B: "qwerty"}

	var buff bytes.Buffer
	buff.Write(bcs.MustMarshal(&s))

	d := bcs.NewDecoder(&buff)
	var s2 TestStruct
	d.Decode(&s2)
	require.NoError(t, d.Err())

	// now we will try to decode again, but buffer is empty
	d.Decode(&s2)
	require.Error(t, d.Err())

	// and even if we write something to buffer, it should not help
	buff.Write(bcs.MustMarshal(&s))
	d.Decode(&s2)
	require.Error(t, d.Err())
}

type FunkyStruct struct {
	A            uint64
	CanSerialize bool   `bcs:"-"`
	ErrorMessage string `bcs:"-"`
	MissError    bool   `bcs:"-"`
}

func (s *FunkyStruct) MarshalBCS(e *bcs.Encoder) error {
	if !s.CanSerialize {
		err := fmt.Errorf("test error from FunkyStruct")
		if s.ErrorMessage != "" {
			err = errors.New(s.ErrorMessage)
		}

		if s.MissError {
			e.Encode(&FunkyStruct{
				CanSerialize: false,
				ErrorMessage: err.Error() + " (missed)",
			})

			return e.Err()
		}

		return err
	}

	e.Encode(s.A)
	return nil
}

func (s *FunkyStruct) UnmarshalBCS(d *bcs.Decoder) error {
	if !s.CanSerialize {
		err := fmt.Errorf("test error from FunkyStruct")
		if s.ErrorMessage != "" {
			err = errors.New(s.ErrorMessage)
		}

		if s.MissError {
			d.Decode(&FunkyStruct{
				CanSerialize: false,
				ErrorMessage: err.Error() + " (missed)",
			})

			return d.Err()
		}

		return err
	}

	s.A = d.ReadUint64()
	return nil
}

func TestEncodingDefferedErrorHandling(t *testing.T) {
	var buff bytes.Buffer
	e := bcs.NewEncoder(&buff)

	e.Encode(&FunkyStruct{
		A:            42,
		CanSerialize: true,
	})
	require.NoError(t, e.Err())

	// Writing something which will result in error
	e.Encode(&FunkyStruct{
		A:            42,
		CanSerialize: false,
	})
	require.Error(t, e.Err())

	// Trying again to write something good
	e.Encode(&FunkyStruct{
		A:            42,
		CanSerialize: true,
	})
	require.Error(t, e.Err())

	// Something else good
	e.WriteInt(123)
	require.Error(t, e.Err())
}

type StructWithFunkyFields struct {
	a  string
	f1 FunkyStruct
	f2 FunkyStruct
	b  int
}

func (s *StructWithFunkyFields) MarshalBCS(e *bcs.Encoder) error {
	e.Encode(s.a)
	e.Encode(s.f1)
	e.Encode(s.f2)
	e.Encode(s.b)

	return nil // error will be checked automatically
}

func (s *StructWithFunkyFields) UnmarshalBCS(d *bcs.Decoder) error {
	s.a = d.ReadString()
	d.Decode(&s.f1)
	d.Decode(&s.f2)
	s.b = d.ReadInt()

	return nil // error will be checked automatically
}

func TestAutomaticErrorCheck(t *testing.T) {
	// First check good case
	encoded := bcs.MustMarshal(&StructWithFunkyFields{
		a: "qwerty",
		f1: FunkyStruct{
			A:            42,
			CanSerialize: true,
		},
		f2: FunkyStruct{
			A:            56,
			CanSerialize: true,
		},
		b: 123,
	})

	decoded := bcs.MustUnmarshalInto(encoded, &StructWithFunkyFields{
		f1: FunkyStruct{
			CanSerialize: true,
		},
		f2: FunkyStruct{
			CanSerialize: true,
		},
	})
	require.Equal(t, "qwerty", decoded.a)
	require.Equal(t, uint64(42), decoded.f1.A)
	require.Equal(t, 123, decoded.b)

	// Now check error case
	_, err := bcs.Marshal(&StructWithFunkyFields{
		a: "qwerty",
		f1: FunkyStruct{
			A:            42,
			CanSerialize: false,
		},
		f2: FunkyStruct{
			A:            56,
			CanSerialize: false,
		},
		b: 123,
	})
	require.Error(t, err)
	expectedErr := "encoding *bcs_test.StructWithFunkyFields: *bcs_test.StructWithFunkyFields: custom encoder: encoding bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())

	// Also check, that second error won't overwrite first one
	_, err = bcs.Marshal(&StructWithFunkyFields{
		a: "qwerty",
		f1: FunkyStruct{
			A:            42,
			CanSerialize: false,
			ErrorMessage: "other error",
		},
		f2: FunkyStruct{
			A:            56,
			CanSerialize: false,
		},
		b: 123,
	})
	require.Error(t, err)
	expectedErr = "encoding *bcs_test.StructWithFunkyFields: *bcs_test.StructWithFunkyFields: custom encoder: encoding bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: other error"
	require.Equal(t, expectedErr, err.Error())

	// And check that second error was also there
	_, err = bcs.Marshal(&StructWithFunkyFields{
		a: "qwerty",
		f1: FunkyStruct{
			A:            42,
			CanSerialize: true,
			ErrorMessage: "other error",
		},
		f2: FunkyStruct{
			A:            56,
			CanSerialize: false,
		},
		b: 123,
	})
	require.Error(t, err)
	expectedErr = "encoding *bcs_test.StructWithFunkyFields: *bcs_test.StructWithFunkyFields: custom encoder: encoding bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())
}

func TestExcessBytesErr(t *testing.T) {
	_, err := bcs.Unmarshal[byte]([]byte{1, 2, 3, 4})
	require.Error(t, err)
	require.Contains(t, err.Error(), "excess bytes: 3")

	_, err = bcs.Unmarshal[string]([]byte{3, 'a', 'b', 'c', 'd', 'e'})
	require.Error(t, err)
	require.Contains(t, err.Error(), "excess bytes: 2")

	_, err = bcs.Unmarshal[BasicStruct]([]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		3, 'a', 'b', 'c', 'd',
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "excess bytes: 1")

	_, err = bcs.Unmarshal[WithByteArr]([]byte{
		3, 'a', 'b', 'c',
		5, 1, 2, 3, 4, 5,
		4, 'e', 'f', 'g', 'h',
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "bytearr: excess bytes: 1")
}

func TestErrorMsg(t *testing.T) {
	type Inner struct {
		A int
		F *FunkyStruct `bcs:"optional"`
		S []*FunkyStruct
		C string
	}

	type Outer struct {
		D     bool
		Inner *Inner
		F     uint32
	}

	vErrInFieldOfField := Outer{
		D: true,
		Inner: &Inner{
			A: 42,
			F: &FunkyStruct{
				A:            42,
				CanSerialize: true,
			},
		},
		F: 123,
	}

	vErrInFieldOfField.Inner.F.CanSerialize = false
	_, err := bcs.Marshal(&vErrInFieldOfField)
	require.Error(t, err)
	expectedErr := "encoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: F: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())

	vErrInFieldOfField.Inner.F.MissError = true
	_, err = bcs.Marshal(&vErrInFieldOfField)
	require.Error(t, err)
	expectedErr = "encoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: F: *bcs_test.FunkyStruct: custom encoder: encoding *bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct (missed)"
	require.Equal(t, expectedErr, err.Error())

	vErrInFieldOfField.Inner.F.CanSerialize = true
	vErrInFieldOfField.Inner.F.MissError = false
	encodedFF := bcs.MustMarshal(&vErrInFieldOfField)

	vErrInFieldOfField.Inner.F.CanSerialize = false
	_, err = bcs.UnmarshalInto(encodedFF, &vErrInFieldOfField)
	require.Error(t, err)
	expectedErr = "decoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: F: bcs_test.FunkyStruct: custom decoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())

	vErrInFieldOfField.Inner.F.MissError = true
	_, err = bcs.UnmarshalInto(encodedFF, &vErrInFieldOfField)
	require.Error(t, err)
	expectedErr = "decoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: F: bcs_test.FunkyStruct: custom decoder: decoding *bcs_test.FunkyStruct: bcs_test.FunkyStruct: custom decoder: test error from FunkyStruct (missed)"
	require.Equal(t, expectedErr, err.Error())

	vErrInElemOfSlice := Outer{
		D: true,
		Inner: &Inner{
			A: 42,
			S: []*FunkyStruct{
				{
					A:            42,
					CanSerialize: true,
				},
			},
		},
		F: 123,
	}

	vErrInElemOfSlice.Inner.S[0].CanSerialize = false
	_, err = bcs.Marshal(&vErrInElemOfSlice)
	require.Error(t, err)
	expectedErr = "encoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: S: []*bcs_test.FunkyStruct: [0]: *bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())

	vErrInElemOfSlice.Inner.S[0].MissError = true
	_, err = bcs.Marshal(&vErrInElemOfSlice)
	require.Error(t, err)
	expectedErr = "encoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: S: []*bcs_test.FunkyStruct: [0]: *bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: encoding *bcs_test.FunkyStruct: *bcs_test.FunkyStruct: custom encoder: test error from FunkyStruct (missed)"
	require.Equal(t, expectedErr, err.Error())

	vErrInElemOfSlice.Inner.S[0].CanSerialize = true
	vErrInElemOfSlice.Inner.S[0].MissError = false
	encodedS := bcs.MustMarshal(&vErrInElemOfSlice)

	vErrInElemOfSlice.Inner.S[0].CanSerialize = false
	_, err = bcs.UnmarshalInto(encodedS, &vErrInElemOfSlice)
	require.Error(t, err)
	expectedErr = "decoding *bcs_test.Outer: bcs_test.Outer: Inner: bcs_test.Inner: S: []*bcs_test.FunkyStruct: [0]: bcs_test.FunkyStruct: custom decoder: test error from FunkyStruct"
	require.Equal(t, expectedErr, err.Error())
}
