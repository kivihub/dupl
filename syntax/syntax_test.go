package syntax

import (
	"github.com/kivihub/dupl/suffixtree"
	"testing"
)

func TestSerialization(t *testing.T) {
	n := genNodes(7)
	n[0].AddChildren(n[1], n[2], n[3])
	n[1].AddChildren(n[4], n[5])
	n[2].AddChildren(n[6])
	m := genNodes(6)
	m[0].AddChildren(m[1], m[2], m[3], m[4], m[5])
	testCases := []struct {
		t        *Node
		expected []int
	}{
		{n[0], []int{6, 2, 0, 0, 1, 0, 0}},
		{m[0], []int{5, 0, 0, 0, 0, 0}},
	}

	for _, tc := range testCases {
		compareSeries(t, Serialize(tc.t), tc.expected)
	}
}

func genNodes(cnt int) []*Node {
	nodes := make([]*Node, cnt)
	for i := range nodes {
		nodes[i] = NewNode()
	}
	return nodes
}

func compareSeries(t *testing.T, stream []*Node, owns []int) {
	if len(stream) != len(owns) {
		t.Errorf("series aren't the same length; got %d, want %d", len(stream), len(owns))
		return
	}
	for i, item := range stream {
		if item.Owns != owns[i] {
			t.Errorf("got %d, want %d", item.Owns, owns[i])
		}
	}
}

func TestGetFuncIndexes(t *testing.T) {
	nodeSeq := "f8 a2 a0 a0 a4 a3 a2 a0 a0 f2 a0 a0"
	nodes := buildFuncNode(nodeSeq)
	testCases := []struct {
		position  suffixtree.Pos
		length    suffixtree.Pos
		threshold int
		expected  []int
	}{
		{0, 12, 3, []int{0, 9}},
		{0, 12, 4, []int{0}},
		{2, 9, 2, []int{0, 9}},
		{2, 9, 3, []int{0}},
	}
Loop:
	for _, tc := range testCases {
		indexes := getFuncIndexes(nodes, tc.position, tc.length, tc.threshold)
		for i := range tc.expected {
			if len(tc.expected) != len(indexes) || tc.expected[i] != indexes[i] {
				t.Errorf("for seq '%s', got %v, want %v", nodeSeq, indexes, tc.expected)
				continue Loop
			}
		}
	}
}

func buildFuncNode(str string) []*Node {
	chars := []rune(str)
	nodes := make([]*Node, (len(chars)+1)/3)
	for i := 0; i < len(chars)-1; i += 3 {
		own := int(chars[i+1] - '0')
		nodeT := int(chars[i])
		if chars[i] == 'f' {
			nodeT = 21 // FuncDecl
		}
		pos := i / 3
		end := pos + own
		nodes[i/3] = &Node{Type: nodeT, Owns: own, Pos: pos, End: end}
	}
	return nodes
}

func TestGetUnitsIndexes(t *testing.T) {
	testCases := []struct {
		seq       string
		threshold int
		expected  []int
	}{
		{"a8 a0 a2 a0 a0", 3, []int{2}},
		{"a0 a8 a2 a0 a0", 1, []int{2}},
		{"a3 a0 a0 a0 a1", 3, []int{0}},
		{"a3 a0 a0 a0 a0", 1, []int{0, 4}},
		{"a1 a0 a1 a0 a0", 2, []int{0, 2}},
	}

Loop:
	for _, tc := range testCases {
		nodes := str2nodes(tc.seq)
		indexes := getUnitsIndexes(nodes, tc.threshold)
		for i := range tc.expected {
			if i > len(indexes)-1 || tc.expected[i] != indexes[i] {
				t.Errorf("for seq '%s', got %v, want %v", tc.seq, indexes, tc.expected)
			}
			continue Loop
		}
	}
}

func TestCyclicDupl(t *testing.T) {
	testCases := []struct {
		seq      string
		indexes  []int
		expected bool
	}{
		{"a1 b0 a2 b0", []int{0, 2}, false},
		{"a1 b0 a1 b0", []int{0, 2}, true},
		{"a0 a0", []int{0, 1}, true},
		{"a1 b0 c1 b0 a1 b0 c1 b0", []int{0, 2, 4, 6}, true},
		{"a1 b0 c1 b0 a1 b0", []int{0, 2, 4}, false},
		{"a0 b0 a0 c0", []int{0, 1, 2, 3}, false},
		{"a0 b0 a0 b0 a0", []int{0, 1, 2}, false},
		{"a1 b0 a1 b0 c1 b0", []int{0, 2, 4}, false},
		{"a1 a1 a1 a1 a1 a1", []int{0, 4}, false},
		{"a2 b0 b0 a2 b0 b0 a2 b0 b0 a2 b0 b0 a2 b0 b0", []int{0, 3, 6, 9, 12}, true},
	}

	for _, tc := range testCases {
		nodes := str2nodes(tc.seq)
		if tc.expected != isCyclic(tc.indexes, nodes) {
			t.Errorf("for seq '%s', indexes %v, got %t, want %t", tc.seq, tc.indexes, !tc.expected, tc.expected)
		}
	}
}

// str2nodes converts strint to a sequence of *Node by following principle:
//   - node is represented by 2 characters
//   - first character is node type
//   - second character is the number for Node.Owns.
func str2nodes(str string) []*Node {
	chars := []rune(str)
	nodes := make([]*Node, (len(chars)+1)/3)
	for i := 0; i < len(chars)-1; i += 3 {
		nodes[i/3] = &Node{Type: int(chars[i]), Owns: int(chars[i+1] - '0')}
	}
	return nodes
}
