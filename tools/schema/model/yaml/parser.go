package yaml

import (
	"strings"
)

func Parse(in []byte) *Node {
	var root Node
	var path []*Node = []*Node{&root} // Nodes in each hierarchy
	var indentList []int = []int{0}   // the list of indent space numbers in the current code block
	var commentNode *Node             // the Node that the current comments block first meets
	cur := &Node{}
	lines := strings.Split(strings.ReplaceAll(string(in), "\r\n", "\n"), "\n")

	var prevIndent, curIndent int = -1, 0
	var comment string
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineNum := i + 1
		next := &Node{}

		var val string
		vals, comments := getComments(lines[i : i+2])
		if strings.TrimSpace(comments[0]) != "" {
			if commentNode == nil {
				commentNode = cur
			}
			comment += ("//" + comments[0][1:] + "\n")

			if commentNode.Line == 0 {
				if strings.TrimSpace(vals[0]) != "" {
					commentNode = cur
				} else if strings.TrimSpace(vals[1]) != "" {
					commentNode = next
				}
			}

			if strings.TrimSpace(comments[1]) == "" {
				// the next line is end of comment block
				commentNode.Comment = comment
				comment = ""
				commentNode = nil
			}

			if strings.TrimSpace(vals[0]) == "" {
				goto end
			}
		}

		val = vals[0]
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
	end:
		cur = next
	}

	return &root
}

func getComments(lines []string) ([]string, []string) {
	val := make([]string, 2)
	comment := make([]string, 2)
	line0 := lines[0]
	idx := strings.Index(line0, "#")
	if idx != -1 {
		val[0], comment[0] = line0[:idx], line0[idx:]
	} else {
		val[0], comment[0] = line0, ""
	}
	line1 := lines[1]
	idx = strings.Index(line1, "#")
	if idx != -1 {
		val[1], comment[1] = line1[:idx], line1[idx:]
	} else {
		val[1], comment[1] = line1, ""
	}
	return val, comment
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
