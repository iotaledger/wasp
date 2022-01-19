package wasmlib

type Proxy struct {
	key     []byte
	kvStore IKvStore
}

func makeKey(key []byte, sep byte, subKey []byte) []byte {
	if len(key) == 0 {
		return subKey
	}
	return append(append(key, sep), subKey...)
}

func subProxy(proxy Proxy, sep byte, subKey []byte) Proxy {
	return Proxy{key: makeKey(proxy.key, sep, subKey), kvStore: proxy.kvStore}
}

func (p Proxy) Buf() *BytesDecoder {
	return NewBytesDecoder(p.Get())
}

func (p Proxy) Delete() {
	p.kvStore.Delete(p.key)
}

func (p Proxy) Exists() bool {
	return p.kvStore.Exists(p.key)
}

func (p Proxy) Get() []byte {
	return p.kvStore.Get(p.key)
}

func (p Proxy) Set(value []byte) {
	p.kvStore.Set(p.key, value)
}

type ArrayProxy struct {
	proxy Proxy
}

func (a ArrayProxy) NewProxy(index uint32) Proxy {
	subKey := NewBytesEncoder().Uint32(index).Data()
	return subProxy(a.proxy, '*', subKey)
}

func (a ArrayProxy) Clear() {
	a.proxy.Set([]byte{0})
}

func (a ArrayProxy) Length() uint32 {
	return a.proxy.Buf().Uint32()
}

type ItemProxy struct {
	proxy Proxy
}

func (a ItemProxy) Clear() {
	a.proxy.Set([]byte{0})
}

func (a ItemProxy) Length() uint32 {
	return a.proxy.Buf().Uint32()
}

type MapProxy struct {
	proxy Proxy
}

func (m MapProxy) NewProxy(subKey []byte) Proxy {
	return subProxy(m.proxy, '.', subKey)
}

func (m MapProxy) Clear() {
	m.proxy.Set([]byte{0})
}

func (m MapProxy) Length() uint32 {
	return m.proxy.Buf().Uint32()
}
