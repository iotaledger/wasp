package yaml

import (
	"strings"
)

func Parse(in []byte) *Node {
	var root Node
	var path []*Node = []*Node{&root} // Nodes in each hierarchy
	var indentList []int = []int{0}   // the list of indent space numbers in the current code block
	var commentNode *Node             // the Node that the current comments block first meets
	var pureComment bool              // whether previous consecutive lines contain only comments

	lines := strings.Split(strings.ReplaceAll(string(in), "\r\n", "\n"), "\n")

	var prevIndent, curIndent int = -1, 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			commentNode = nil
			continue
		}
		lineNum := i + 1
		cur := Node{}

		val, comment := getComment(line)
		if comment != "" {
			if commentNode == nil {
				commentNode = &cur
			}

			commentNode.Comment += (comment + "\n")
			if strings.TrimSpace(val) == "" {
				pureComment = true
				goto next
			} else if pureComment {
				pureComment = false
				if commentNode.Line == 0 {
					// a series of comments meet yaml item first time
					cur.Comment = commentNode.Comment
					commentNode = &cur
				}
			}
		} else {
			pureComment = false
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
				cur.Val = elts[0]
				cur.Contents = append(cur.Contents,
					&Node{
						Val:  strings.TrimSpace(elts[1]),
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
			parent.Contents = append(parent.Contents, &cur)
			path = append(path, &cur)

		} else if curIndent == prevIndent {
			// sibling
			parent := path[len(path)-2]
			parent.Contents = append(parent.Contents, &cur)
			path[len(path)-1] = &cur
		} else {
			// uncle
			path = path[:getLevel(curIndent, indentList)+1] // pop until parent of uncle Node
			uncle := path[len(path)-1]
			uncle.Contents = append(uncle.Contents, &cur)
			path = append(path, &cur)
		}

	next:
		prevIndent = curIndent
	}

	return &root
}

func getComment(in string) (string, string) {
	if strings.Contains(in, "#") {
		elts := strings.SplitN(in, "#", 2)
		if len(elts) == 1 {
			return "", elts[0]
		}
		return elts[0], elts[1]
	}
	return in, ""
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
