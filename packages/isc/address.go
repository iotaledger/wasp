package isc

/*const addressIsNil rwutil.Kind = 0x80

func AddressFromReader(rr *rwutil.Reader) (address cryptolib.Address) {
	kind := rr.ReadKind()
	if kind == addressIsNil {
		return nil
	}
	addrSize := 0
	if rr.Err == nil {
		address, rr.Err = iotago.AddressSelector(uint32(kind))
		if rr.Err != nil {
			addrSize = 0
		} else {
			addrSize = address.Size()
		}
	}
	rr.PushBack().WriteKind(kind)
	rr.ReadSerialized(address, math.MaxUint16, addrSize)
	return address
}

func AddressToWriter(ww *rwutil.Writer, address cryptolib.Address) {
	if address == nil {
		ww.WriteKind(addressIsNil)
		return
	}
	ww.WriteSerialized(address, math.MaxUint16, address.Size())
}

func AddressFromBytes(data []byte) (cryptolib.Address, error) {
	rr := rwutil.NewBytesReader(data)
	return AddressFromReader(rr), rr.Err
}

func AddressToBytes(address cryptolib.Address) []byte {
	ww := rwutil.NewBytesWriter()
	AddressToWriter(ww, address)
	return ww.Bytes()
}*/
