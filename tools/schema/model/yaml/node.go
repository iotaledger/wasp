package yaml

import "encoding/json"

type Node struct {
	Val      string  `json:"val"`
	Line     int     `json:"line"`
	Comment  string  `json:"comment"`
	Contents []*Node `json:"contents"`
}

func (n *Node) String() string {
	out, _ := json.Marshal(n)
	return string(out)
}
