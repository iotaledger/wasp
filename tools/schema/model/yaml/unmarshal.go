package yaml

import (
	"fmt"

	"github.com/iotaledger/wasp/tools/schema/model"
)

func Unmarshal(in []byte, def *model.SchemaDef) error {
	root := Parse(in)
	if root == nil {
		return fmt.Errorf("root is nil")
	}
	return Convert(root, def)
}
