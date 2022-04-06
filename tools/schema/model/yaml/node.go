// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package yaml

import "encoding/json"

type Node struct {
	Val         string  `json:"val"`
	Line        int     `json:"line"`
	HeadComment string  `json:"head_comment"`
	LineComment string  `json:"line_comment"`
	Contents    []*Node `json:"contents"`
}

func (n *Node) String() string {
	out, _ := json.Marshal(n)
	return string(out)
}
