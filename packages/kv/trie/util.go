package trie

func commonPrefix(b1, b2 []byte) []byte {
	ret := make([]byte, 0)
	for i := 0; i < len(b1) && i < len(b2); i++ {
		if b1[i] != b2[i] {
			break
		}
		ret = append(ret, b1[i])
	}
	return ret
}

func assert(cond bool, err interface{}) {
	if !cond {
		panic(err)
	}
}
