package syntax

import (
	"fmt"
	"github.com/kivihub/dupl/suffixtree"
	"log"
)

func FindFuncUnits(data []*Node, m suffixtree.Match, tokenThreshold, funcThreshold int, verbose bool) []Match {
	if len(m.Ps) == 0 {
		return nil
	}

	indexesMap := make(map[int][]int)
	// m.Ps是多个重复Ast的起始位置，如abcdecde，其中cd重复，那么m.Ps就是[2,5]数组
	for i, pos := range m.Ps {
		// indexes是一个重复片段内的多个函数的起始节点列表
		indexes := getFuncIndexes(data, pos, m.Len, tokenThreshold, funcThreshold, verbose)
		if len(indexes) == 0 {
			return nil
		}
		indexesMap[i] = indexes
	}

	// 当重复行数从重复代码段中解析函数时，可能由于行数不同导致解析出不同的函数个数，这种Case则忽略
	// 如重复片段1中的函数A不足20行，于此同时重复片段2中的函数B与函数A的内容相同，只是在其中增加了空行或者注释，最终使行数超过20行
	// 这导致重复片段1解析的函数不包含函数A，而重复片段2解析的函数包含函数B。
	duplFuncGroupNum := len(indexesMap[0])
	for _, indexes := range indexesMap {
		if len(indexes) != duplFuncGroupNum {
			msg := ""
			for _, ints := range indexesMap {
				for _, i3 := range ints {
					msg += fmt.Sprintf("%s:%d\n", data[i3].Filename, data[i3].StartLine)
				}
				msg += "\n"
			}
			fmt.Printf(fmt.Sprintf("Warn found different function number in match, Detail:\n%s", msg))
			return nil
		}
	}

	duplFuncNumPerGroup := len(indexesMap)
	matchs := make([]Match, duplFuncGroupNum)
	for groupIndex := 0; groupIndex < duplFuncGroupNum; groupIndex++ {
		matchs[groupIndex] = Match{Frags: make([][]*Node, duplFuncNumPerGroup)}

		for funcIndex := 0; funcIndex < duplFuncNumPerGroup; funcIndex++ {
			matchs[groupIndex].Frags[funcIndex] = make([]*Node, 1)
			matchs[groupIndex].Frags[funcIndex][0] = data[indexesMap[funcIndex][groupIndex]]
		}
		funcNode := matchs[groupIndex].Frags[0][0]
		if funcNode.EndAtSuffixTree < funcNode.StartAtSuffixTree {
			fmt.Printf("a")
		}
		matchs[groupIndex].Hash = hashSeq(data[funcNode.StartAtSuffixTree:funcNode.EndAtSuffixTree]) // Hash为重复度组的分组标识
	}

	return matchs
}

// nodeSeq解析为多个重复的函数段，并过滤小于函数阈值的函数
func getFuncIndexes(data []*Node, position, length suffixtree.Pos, tokenThreshold, funcThreshold int, verbose bool) []int {
	var indexes []int

	nodeSeq := data[position : position+length]
	duplFragPos := int(position)
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
		duplLastNodeIndexAtNodeSeq, lastLine := getFuncDuplLastNodeIndexAndLine(nodeSeq, i, funcNode)
		if duplLastNodeIndexAtNodeSeq == -1 || lastLine == -1 {
			panic("Unexpect error at getFuncDuplLastNodeIndexAndLine")
		}
		dupLines := lastLine - n.StartLine + 1
		duplLastNodeIndex := duplLastNodeIndexAtNodeSeq + int(position)

		// 3. 超过阈值则加入{indexes}
		if duplLastNodeIndex-duplFragPos >= tokenThreshold && dupLines >= funcThreshold {
			if verbose {
				log.Printf("duplicate lines %s:%d-%d\n", funcNode.Filename, n.StartLine, lastLine)
			}
			indexes = append(indexes, funcNodeIndex)
			funcNode.StartAtSuffixTree = funcNodeIndex
			funcNode.EndAtSuffixTree = duplLastNodeIndex
			GlobalFuncDuplManager.AddDuplFrag(funcNode, n.StartLine, lastLine)
		}

		// 4. i = FuncEnd Index + 1
		duplFragPos = duplLastNodeIndex + 1
		i = duplLastNodeIndexAtNodeSeq + 1
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
