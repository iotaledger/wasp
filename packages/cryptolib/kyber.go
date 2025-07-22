package cryptolib

import (
	"go.dedis.ch/kyber/v3"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

func PointFromBytes(data []byte, factory interface{ Point() kyber.Point }) (point kyber.Point, err error) {
	point = factory.Point()
	err = point.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return point, nil
}

func PointFromReader(rr *rwutil.Reader, factory interface{ Point() kyber.Point }) (point kyber.Point) {
	point = factory.Point()
	rr.ReadFromFunc(point.UnmarshalFrom)
	return point
}

func PointToWriter(ww *rwutil.Writer, point kyber.Point) {
	ww.WriteFromFunc(point.MarshalTo)
}

func ScalarFromBytes(data []byte, factory interface{ Scalar() kyber.Scalar }) (scalar kyber.Scalar, err error) {
	scalar = factory.Scalar()
	err = scalar.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return scalar, nil
}

func ScalarFromReader(rr *rwutil.Reader, factory interface{ Scalar() kyber.Scalar }) (scalar kyber.Scalar) {
	scalar = factory.Scalar()
	rr.ReadFromFunc(scalar.UnmarshalFrom)
	return scalar
}

func ScalarToWriter(ww *rwutil.Writer, scalar kyber.Scalar) {
	ww.WriteFromFunc(scalar.MarshalTo)
}
