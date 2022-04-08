// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"strings"
)

func Parse(in []byte) *Node {
	var root Node
	path := []*Node{&root} // Nodes in each hierarchy
	indentList := []int{0} // the list of indent space numbers in the current code block
	lines := strings.Split(strings.ReplaceAll(string(in), "\r\n", "\n"), "\n")

	prevIndent, curIndent := -1, 0
	var comment string
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			comment = ""
			continue
		}
		lineNum := i + 1
		cur := &Node{}

		val, oriComment := getComment(line)
		if strings.TrimSpace(val) != "" {
			if comment != "" {
				cur.HeadComment = comment
			} else if strings.TrimSpace(oriComment) != "" {
				cur.LineComment = ("//" + oriComment[1:] + "\n")
			}
			comment = ""
		} else {
			comment += ("//" + oriComment[1:] + "\n")
			if strings.TrimSpace(oriComment) == "" {
				comment = ""
			}
			continue
		}

		cur.Line = lineNum
		curIndent = len(val) - len(strings.TrimLeft(val, " "))
		val = strings.TrimSpace(val)
		if strings.Contains(val, ":") {
			if val[len(val)-1] == ':' {
				// yaml tag only
				cur.Val = val[:len(val)-1]
			} else {
				// yaml tag with value
				elts := strings.Split(val, ":")
				cur.Val = strings.TrimSpace(elts[0])
				value := strings.TrimSpace(elts[1])
				// TODO note that value string could be quoted! What if there is only one quote?
				// what about special characters in value string?
				if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				cur.Contents = append(cur.Contents,
					&Node{
						Val:  value,
						Line: lineNum,
					})
			}
		} else {
			// yaml value only
			cur.Val = val
		}

		if curIndent > prevIndent {
			// child
			indentList = setIndentList(curIndent, indentList)
			if indentList == nil {
				// style tweak
				return nil
			}
			parent := path[len(path)-1]
			parent.Contents = append(parent.Contents, cur)
			path = append(path, cur)
		} else if curIndent == prevIndent {
			// sibling
			parent := path[len(path)-2]
			parent.Contents = append(parent.Contents, cur)
			path[len(path)-1] = cur
		} else {
			// uncle
			path = path[:getLevel(curIndent, indentList)+1] // pop until parent of uncle Node
			uncle := path[len(path)-1]
			uncle.Contents = append(uncle.Contents, cur)
			path = append(path, cur)
		}

		prevIndent = curIndent
	}

	return &root
}

func getComment(line string) (string, string) {
	idx := strings.Index(line, "#")
	if idx != -1 {
		return line[:idx], line[idx:]
	}
	return line, ""
}

func getLevel(indent int, path []int) int {
	for i, ind := range path {
		if indent == ind {
			return i
		}
	}
	return 0
}

func setIndentList(indent int, list []int) []int {
	if list[len(list)-1] < indent {
		return append(list, indent)
	}
	for _, elt := range list {
		if elt == indent {
			return list
		}
	}
	return nil
}
