package wasmlib

type Proxy struct {
	key     []byte
	kvStore IKvStore
}

var _ IKvStore = new(Proxy)

func (p Proxy) Buf(key []byte) *BytesDecoder {
	return NewBytesDecoder(p.Get(key))
}

func (p Proxy) Delete(key []byte) {
	p.kvStore.Delete(key)
}

func (p Proxy) Exists(key []byte) bool {
	return p.kvStore.Exists(key)
}

func (p Proxy) Get(key []byte) []byte {
	return p.kvStore.Get(key)
}

func (p Proxy) Set(key, value []byte) {
	p.kvStore.Set(key, value)
}

type ArrayProxy struct {
	proxy Proxy
}

func (a ArrayProxy) Clear() {
	a.proxy.Set(a.proxy.key, []byte{0})
}

func (a ArrayProxy) Length() uint32 {
	return a.proxy.Buf(a.proxy.key).Uint32()
}

type ItemProxy struct {
	proxy Proxy
}

func (a ItemProxy) Clear() {
	a.proxy.Set(a.proxy.key, []byte{0})
}

func (a ItemProxy) Length() uint32 {
	return a.proxy.Buf(a.proxy.key).Uint32()
}

type MapProxy struct {
	proxy Proxy
}

func (m MapProxy) Clear() {
	m.proxy.Set(m.proxy.key, []byte{0})
}

func (m MapProxy) Length() uint32 {
	return m.proxy.Buf(m.proxy.key).Uint32()
}
