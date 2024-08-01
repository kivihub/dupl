package syntax

import (
	"fmt"
	"github.com/kivihub/dupl/utils"
	"log"
)

var GlobalFuncDuplManager *FuncDuplManager

func InitFuncDuplManager(funcDuplRatio int, verbose bool) {
	GlobalFuncDuplManager = &FuncDuplManager{
		verbose:       verbose,
		funcDuplRatio: funcDuplRatio,
		funcDuplFrags: make(map[string]*FuncDupl),
	}
}

type FuncDuplManager struct {
	verbose       bool
	funcDuplRatio int
	funcDuplFrags map[string]*FuncDupl // key为file:funcPosition
}

type FuncDupl struct {
	FuncNode  *Node
	DuplFrags []*DuplFrag
}

type DuplFrag struct {
	StartLine int
	EndLine   int
}

func (m *FuncDuplManager) BothFuncNeedExceedRatio() bool {
	return m.funcDuplRatio > 0
}

func (m *FuncDuplManager) AddDuplFrag(funcNode *Node, duplStartLine, duplEndLine int) {
	if m.funcDuplRatio == 0 {
		return
	}

	key := nodeKey(funcNode)
	duplFrag := &DuplFrag{
		StartLine: duplStartLine,
		EndLine:   duplEndLine,
	}
	if v, exist := m.funcDuplFrags[key]; !exist {
		v := &FuncDupl{
			FuncNode:  funcNode,
			DuplFrags: []*DuplFrag{duplFrag},
		}
		m.funcDuplFrags[key] = v
	} else {
		v.DuplFrags = append(v.DuplFrags, duplFrag)
	}
}

func (m *FuncDuplManager) RemoveFuncLessRatio() {
	if m.funcDuplRatio == 0 {
		return
	}
	for key, dupl := range m.funcDuplFrags {
		funcNode := dupl.FuncNode
		funcLines := funcNode.EndLine - funcNode.StartLine + 1
		totalDuplLines := 0
		dupls := make([]string, 0)
		for _, frag := range dupl.DuplFrags {
			duplLines := frag.EndLine - frag.StartLine + 1
			totalDuplLines += duplLines
			dupls = append(dupls, fmt.Sprintf("%d-%d", frag.StartLine, frag.EndLine))
		}
		realRatio := 100 * totalDuplLines / funcLines
		if m.verbose {
			log.Printf("Function:%s DuplRatio:%d%% DuplFrags:%v\n", nodeKey(funcNode), realRatio, utils.MarshalPretty(dupls))
		}
		if realRatio < utils.IntAbs(m.funcDuplRatio) {
			delete(m.funcDuplFrags, key)
		}
	}
}

func (m *FuncDuplManager) Exist(fileName string, funcPos int) (exist bool) {
	if m.funcDuplRatio == 0 {
		exist = true
	} else {
		key := fmt.Sprintf("%s:%d", fileName, funcPos)
		_, exist = m.funcDuplFrags[key]
	}
	return
}

func nodeKey(node *Node) string {
	return fmt.Sprintf("%s:%d", node.Filename, node.StartLine)
}
