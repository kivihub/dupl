package syntax

import (
	"crypto/sha1"
	"github.com/kivihub/dupl/suffixtree"
)

type Node struct {
	Type               int
	Filename           string
	Pos, End           int
	StartLine, EndLine int
	Children           []*Node
	Owns               int
	Source             string
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) AddChildren(children ...*Node) {
	n.Children = append(n.Children, children...)
}

func (n *Node) Val() int {
	return n.Type
}

type Match struct {
	Hash  string
	Frags [][]*Node
}

func Serialize(n *Node) []*Node {
	stream := make([]*Node, 0, 10)
	serial(n, &stream)
	return stream
}

func serial(n *Node, stream *[]*Node) int {
	*stream = append(*stream, n)
	var count int
	for _, child := range n.Children {
		count += serial(child, stream)
	}
	n.Owns = count
	return count + 1
}

// FindSyntaxUnits finds all complete syntax units in the match group and returns them
// with the corresponding hash.
func FindSyntaxUnits(data []*Node, m suffixtree.Match, threshold int) Match {
	if len(m.Ps) == 0 {
		return Match{}
	}
	firstSeq := data[m.Ps[0] : m.Ps[0]+m.Len]       // 最后一个元素
	indexes := getUnitsIndexes(firstSeq, threshold) // 计算连续的完整AstNode在{firstSeq}中的Index集合

	// TODO: is this really working?
	indexCnt := len(indexes)
	if indexCnt > 0 {
		lasti := indexes[indexCnt-1] // {firstSeq}中最后完整AstNode的Index
		firstn := firstSeq[lasti]    // {firstSeq}中最后完整AstNode的Index - 对应的Node
		for i := 1; i < len(m.Ps); i++ {
			n := data[int(m.Ps[i])+lasti]
			if firstn.Owns != n.Owns { // 最后一个重复AstNode的Own校验：保证重复两方的AstNode的Children数量相同；
				indexes = indexes[:indexCnt-1]
				break
			}
		}
	}

	// 以下情况则返回不匹配
	// 重复片段由多个小的重复片段拼接组成
	// 重复度内容跨文件
	if len(indexes) == 0 || isCyclic(indexes, firstSeq) || spansMultipleFiles(indexes, firstSeq) {
		return Match{}
	}

	match := Match{Frags: make([][]*Node, len(m.Ps))}
	// m.Ps是多个重复Ast的起始位置，如abcdecde，其中cdf重复，那么m.Ps就是[2,5]数组
	for i, pos := range m.Ps {
		match.Frags[i] = make([]*Node, len(indexes))
		for j, index := range indexes { // index可看做重复长度
			match.Frags[i][j] = data[int(pos)+index]
		}
	}

	lastIndex := indexes[len(indexes)-1]
	match.Hash = hashSeq(firstSeq[indexes[0] : lastIndex+firstSeq[lastIndex].Owns])
	return match
}

// 返回连续的多个AstNode的Index，其中每个AstNode的owns超过阈值threshold
func getUnitsIndexes(nodeSeq []*Node, threshold int) []int {
	var indexes []int
	var split bool // 标识是否连续，如果为true则表示不连续，需要clear {indexes}
	for i := 0; i < len(nodeSeq); {
		n := nodeSeq[i]
		switch {
		case n.Owns >= len(nodeSeq)-i: // 不完整Ast Node
			// not complete syntax unit
			i++
			split = true
			continue
		case n.Owns+1 < threshold: // 完整Ast Node，但是小于阈值
			split = true
		default: // 完整Ast Node，且大于阈值
			if split {
				indexes = indexes[:0]
				split = false
			}
			indexes = append(indexes, i)
		}
		i += n.Owns + 1
	}
	return indexes
}

// isCyclic finds out whether there is a repetive pattern in the found clone. If positive,
// it return false to point out that the clone would be redundant.
func isCyclic(indexes []int, nodes []*Node) bool {
	cnt := len(indexes)
	if cnt <= 1 {
		return false
	}

	alts := make(map[int]bool)
	for i := 1; i <= cnt/2; i++ {
		if cnt%i == 0 {
			alts[i] = true
		}
	}

	for i := 0; i < indexes[cnt/2]; i++ {
		nstart := nodes[i+indexes[0]]
	AltLoop:
		for alt := range alts {
			for j := alt; j < cnt; j += alt {
				index := i + indexes[j]
				if index < len(nodes) {
					nalt := nodes[index]
					if nstart.Owns == nalt.Owns && nstart.Type == nalt.Type {
						continue
					}
				} else if i >= indexes[alt] {
					return true
				}
				delete(alts, alt)
				continue AltLoop
			}
		}
		if len(alts) == 0 {
			return false
		}
	}
	return true
}

func spansMultipleFiles(indexes []int, nodes []*Node) bool {
	if len(indexes) < 2 {
		return false
	}
	f := nodes[indexes[0]].Filename
	for i := 1; i < len(indexes); i++ {
		if nodes[indexes[i]].Filename != f {
			return true
		}
	}
	return false
}

func hashSeq(nodes []*Node) string {
	h := sha1.New()
	bytes := make([]byte, len(nodes))
	for i, node := range nodes {
		bytes[i] = byte(node.Type)
	}
	h.Write(bytes)
	return string(h.Sum(nil))
}
