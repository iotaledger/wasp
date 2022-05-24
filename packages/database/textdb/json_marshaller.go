package textdb

import "encoding/json"

type jsonMarshaller struct{}

var _ marshaller = &jsonMarshaller{}

func (m *jsonMarshaller) marshal(val interface{}) ([]byte, error) {
	return json.MarshalIndent(val, "", " ")
}

func (m *jsonMarshaller) unmarshal(buf []byte, v interface{}) error {
	return json.Unmarshal(buf, v)
}
