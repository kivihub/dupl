package syntax

import (
	"github.com/kivihub/dupl/suffixtree"
	"log"
)

func FindFuncUnits(data []*Node, m suffixtree.Match, funcThreshold int, verbose bool) Match {
	if len(m.Ps) == 0 {
		return Match{}
	}

	match := Match{Frags: make([][]*Node, len(m.Ps))}
	// m.Ps是多个重复Ast的起始位置，如abcdecde，其中cdf重复，那么m.Ps就是[2,5]数组
	for i, pos := range m.Ps {
		indexes := getFuncIndexes(data, pos, m.Len, funcThreshold, verbose)
		if len(indexes) == 0 {
			return Match{}
		}

		match.Frags[i] = make([]*Node, len(indexes))
		for j, index := range indexes {
			match.Frags[i][j] = data[index]
		}
	}

	match.Hash = hashSeq(data[m.Ps[0] : m.Ps[0]+m.Len]) // Hash为重复度组的分组标识
	return match
}

// nodeSeq解析为多个重复的函数段，并过滤小于函数阈值的函数
func getFuncIndexes(data []*Node, position, length suffixtree.Pos, funcThreshold int, verbose bool) []int {
	var indexes []int

	nodeSeq := data[position : position+length]
	for i := 0; i < len(nodeSeq); {
		n := nodeSeq[i]
		// 1. 获取结点所在的函数起始行
		funcNodeIndex := findBelongsFuncNode(n, data, position+suffixtree.Pos(i))
		if funcNodeIndex == -1 { // Node不在函数内
			i++
			continue
		}
		funcNode := data[funcNodeIndex]

		// 2. 获取函数的重复行数
		duplLastNodeIndex, lastLine := getFuncDuplLastNodeIndexAndLine(nodeSeq, i, funcNode)
		if duplLastNodeIndex == -1 || lastLine == -1 {
			panic("Unexpect error at getFuncDuplLastNodeIndexAndLine")
		}
		dupLines := lastLine - n.StartLine + 1

		// 3. 超过阈值则加入{indexes}
		if dupLines >= funcThreshold {
			if verbose {
				log.Printf("duplicate lines %s:%d-%d\n", funcNode.Filename, n.StartLine, lastLine)
			}
			indexes = append(indexes, funcNodeIndex)
			GlobalFuncDuplManager.AddDuplFrag(funcNode, n.StartLine, lastLine)
			return indexes
		}

		// 4. i = FuncEnd Index + 1
		i = duplLastNodeIndex + 1
	}
	return indexes
}

func getFuncDuplLastNodeIndexAndLine(nodeSeq []*Node, i int, funcNode *Node) (int, int) {
	lastLine := -1
	lastIndex := -1
	for i < len(nodeSeq) {
		cur := nodeSeq[i]
		if cur.Filename != funcNode.Filename || cur.Pos > funcNode.End {
			break
		}

		if cur.Owns < len(nodeSeq)-i { // 完整语法块在重复段内
			lastIndex = i + cur.Owns
			if cur.EndLine > lastLine {
				lastLine = cur.EndLine
			}
		} else {
			lastIndex = i
			if cur.StartLine > lastLine {
				lastLine = cur.StartLine
			}
		}
		i = lastIndex + 1
	}
	return lastIndex, lastLine
}

func findBelongsFuncNode(node *Node, data []*Node, position suffixtree.Pos) int {
	for i := position; i >= 0; i-- {
		cur := data[i]
		if cur.Type == 21 { // FuncDecl is 21, use magic number for avoid cycle import
			if cur.Filename == node.Filename && cur.End >= node.Pos {
				return int(i)
			} else {
				return -1
			}
		}
	}
	return -1
}
