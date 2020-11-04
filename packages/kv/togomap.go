package kv

func ToGoMap(kv KVStore) map[Key][]byte {
	r := make(map[Key][]byte)
	kv.Iterate(EmptyPrefix, func(k Key, v []byte) bool {
		r[k] = v
		return true
	})
	return r
}
