// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"errors"

	"github.com/iotaledger/wasp/tools/schema/model"
)

func Unmarshal(in []byte, def *model.SchemaDef) error {
	root := Parse(in)
	if root == nil {
		return errors.New("failed to parse input yaml file")
	}
	return Convert(root, def)
}
