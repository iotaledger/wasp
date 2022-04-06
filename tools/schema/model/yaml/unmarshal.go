// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"

	"github.com/iotaledger/wasp/tools/schema/model"
)

func Unmarshal(in []byte, def *model.SchemaDef) error {
	root := Parse(in)
	if root == nil {
		return fmt.Errorf("failed to parse input yaml file")
	}
	return Convert(root, def)
}
