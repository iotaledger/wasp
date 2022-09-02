package crypto

import (
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
)

// PointUnmarshalFrom decodes the content of the point P by reading from an io.Reader.
// It performs all additional sanity checks, if present.
func PointUnmarshalFrom(P kyber.Point, r io.Reader) (int, error) { //nolint:gocritic
	data := make([]byte, P.MarshalSize())
	n, err := io.ReadFull(r, data)
	if err != nil {
		return n, fmt.Errorf("reading failed: %w", err)
	}
	if err := P.UnmarshalBinary(data); err != nil {
		return n, fmt.Errorf("parsing failed: %w", err)
	}

	if i, ok := P.(interface{ IsCanonical(b []byte) bool }); ok && !i.IsCanonical(data) {
		return n, ErrNotCanonical
	}
	if i, ok := P.(interface{ HasSmallOrder() bool }); ok && i.HasSmallOrder() {
		return n, ErrSmallOrder
	}
	return n, nil
}

// ScalarUnmarshalFrom decodes the content of the scalar s by reading from an io.Reader.
// It performs all additional sanity checks, if present.
func ScalarUnmarshalFrom(s kyber.Scalar, r io.Reader) (int, error) {
	data := make([]byte, s.MarshalSize())
	n, err := io.ReadFull(r, data)
	if err != nil {
		return n, fmt.Errorf("reading failed: %w", err)
	}
	if err := scalarUnmarshalBinary(s, data); err != nil {
		return n, err
	}
	return n, nil
}

func scalarUnmarshalBinary(s kyber.Scalar, data []byte) error {
	if err := s.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	if i, ok := s.(interface{ IsCanonical(b []byte) bool }); ok && !i.IsCanonical(data) {
		return ErrNotCanonical
	}
	return nil
}
