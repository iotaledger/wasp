package errors

/*
func ErrorFromBytes(mu *marshalutil.MarshalUtil) (*errors.BlockError, error) {
	var err error
	var hash uint32
	//var prefixId uint32
	var errorId uint32
	var blockError *BlockError

	if isError, err := mu.ReadBool(); err != nil {
		return nil, err
	} else if !isError {
		return nil, nil
	}

	if _, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if errorId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if blockError, err = e.Create(errorId, int(errorId)); err != nil {
		return nil, err
	}

	if hash, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if err = blockError.deserializeParams(mu); err != nil {
		return nil, err
	}

	newHash := blockError.Hash()

	if newHash != hash {
		return nil, xerrors.Errorf("Hash of supplied error does not match the serialized form! %v:%v", hash, newHash)
	}

	return blockError, nil
}
*/
