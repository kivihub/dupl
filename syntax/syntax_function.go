package syntax

import (
	"github.com/kivihub/dupl/suffixtree"
	"log"
)

func FindFuncUnits(data []*Node, m suffixtree.Match, funcThreshold int) Match {
	match := Match{Frags: make([][]*Node, len(m.Ps))}
	// m.Ps是多个重复Ast的起始位置，如abcdecde，其中cdf重复，那么m.Ps就是[2,5]数组
	for i, pos := range m.Ps {
		indexes := getFuncIndexes(data, pos, m.Len, funcThreshold)
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
func getFuncIndexes(data []*Node, position, length suffixtree.Pos, funcThreshold int) []int {
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
		duplLastNodeIndex := getFuncDuplLastNodeIndex(nodeSeq, i, funcNode)
		duplLastNode := nodeSeq[duplLastNodeIndex]
		dupLines := duplLastNode.StartLine - n.StartLine + 1

		// 3. 超过阈值则加入{indexes}
		if dupLines >= funcThreshold {
			log.Printf("duplicate lines %s:%d-%d", funcNode.Filename, n.StartLine, duplLastNode.StartLine)
			indexes = append(indexes, funcNodeIndex)
			GlobalFuncDuplManager.AddDuplFrag(funcNode, n.StartLine, duplLastNode.StartLine)
		}

		// 4. i = FuncEnd Index + 1
		i = duplLastNodeIndex + 1
	}
	return indexes
}

func getFuncDuplLastNodeIndex(nodeSeq []*Node, i int, node *Node) int {
	lastIndex := i
	for i++; i < len(nodeSeq); i++ {
		cur := nodeSeq[i]
		if cur.Filename != node.Filename || cur.Pos > node.End {
			break
		} else {
			lastIndex = i
		}
	}
	return lastIndex
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
