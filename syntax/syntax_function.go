package syntax

import (
	"github.com/kivihub/dupl/suffixtree"
	"log"
)

func FindFuncUnits(data []*Node, m suffixtree.Match, funcThreshold, funcRatio int) Match {
	match := Match{Frags: make([][]*Node, len(m.Ps))}
	// m.Ps是多个重复Ast的起始位置，如abcdecde，其中cdf重复，那么m.Ps就是[2,5]数组
	exceedFuncRatio := false
	for i, pos := range m.Ps {
		indexes, ratios := getFuncIndexes(data, pos, m.Len, funcThreshold)
		if len(indexes) == 0 {
			return Match{}
		}
		exceedFuncRatio = exceedFuncRatio || isAnyBigger(ratios, funcRatio)

		match.Frags[i] = make([]*Node, len(indexes))
		for j, index := range indexes {
			match.Frags[i][j] = data[index]
		}
	}

	if !exceedFuncRatio { // 如果不存在任意一个大于{funcRatio}的重复段，则返回空
		return Match{}
	}

	match.Hash = hashSeq(data[m.Ps[0] : m.Ps[0]+m.Len]) // not work nicely, should unique duplicate pair while printing
	return match
}

func isAnyBigger(arr []int, v int) bool {
	for _, item := range arr {
		if item >= v {
			return true
		}
	}
	return false
}

// nodeSeq解析为多个重复的函数段，并过滤小于函数阈值的函数
func getFuncIndexes(data []*Node, position, length suffixtree.Pos, funcThreshold int) ([]int, []int) {
	var indexes, ratios []int

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
		funcLines := funcNode.EndLine - funcNode.StartLine + 1
		ratio := 100 * dupLines / funcLines
		if dupLines >= funcThreshold {
			log.Printf("duplicate lines %s:%d-%d\tratio:%d/%d=%d%%", funcNode.Filename, n.StartLine, duplLastNode.StartLine, dupLines, funcLines, ratio)
			indexes = append(indexes, funcNodeIndex)
			ratios = append(ratios, ratio)
		}

		// 4. i = FuncEnd Index + 1
		i = duplLastNodeIndex + 1
	}
	return indexes, ratios
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
