package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type PublicChainMetadata struct {
	EVMJsonRPCURL   string
	EVMWebSocketURL string
	Name            string
	Description     string
	Website         string
}

func PublicChainMetadataFromBytes(data []byte) (*PublicChainMetadata, error) {
	return rwutil.ReadFromBytes(data, new(PublicChainMetadata))
}

func (m *PublicChainMetadata) Bytes() []byte {
	return rwutil.WriteToBytes(m)
}

func (m *PublicChainMetadata) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.EVMJsonRPCURL = rr.ReadString()
	m.EVMWebSocketURL = rr.ReadString()
	m.Name = rr.ReadString()
	m.Description = rr.ReadString()
	m.Website = rr.ReadString()
	return rr.Err
}

func (m *PublicChainMetadata) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteString(m.EVMJsonRPCURL)
	ww.WriteString(m.EVMWebSocketURL)
	ww.WriteString(m.Name)
	ww.WriteString(m.Description)
	ww.WriteString(m.Website)
	return ww.Err
}
