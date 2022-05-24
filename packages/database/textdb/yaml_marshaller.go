package textdb

import "gopkg.in/yaml.v2"

type yamlMarshaller struct{}

var _ marshaller = &yamlMarshaller{}

func (y *yamlMarshaller) marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y *yamlMarshaller) unmarshal(in []byte, v interface{}) error {
	return yaml.Unmarshal(in, v)
}
