package suffixtree

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

const infinity = math.MaxInt32

// Pos denotes position in data slice.
type Pos int32

type Token interface {
	Val() int
}

// STree is a struct representing a suffix tree.
type STree struct {
	data     []Token // AST的Type, 有对应的[]Node变量，见buildtree.go#BuildTree
	root     *state  // 根节点
	auxState *state  // auxiliary state，根节点的{后缀连接}

	// active point，活动节点
	s          *state
	start, end Pos
}

// New creates new suffix tree.
func New() *STree {
	t := new(STree)
	t.data = make([]Token, 0, 50)
	t.root = newState(t)
	t.auxState = newState(t)
	t.root.linkState = t.auxState
	t.s = t.root
	return t
}

// Update refreshes the suffix tree to by new data.
func (t *STree) Update(data ...Token) {
	t.data = append(t.data, data...)
	for range data {
		t.update()
		t.s, t.start = t.canonize(t.s, t.start, t.end)
		t.end++
	}
}

// update transforms suffix tree T(n) to T(n+1).
func (t *STree) update() {
	oldr := t.root

	// (s, (start, end)) is the canonical reference pair for the active point
	s := t.s
	start, end := t.start, t.end
	var r *state
	for {
		var endPoint bool

		// 查看新的节点是否已包含在{活动节点}的tran中，如果不包含则分裂{活动节点}的对应tran
		r, endPoint = t.testAndSplit(s, start, end-1)

		// 是否需要新增节点，如果是endPoint则不需要新增，之前的tran已包含
		if endPoint {
			break
		}
		// 增加新的{后缀节点}
		r.fork(end)

		if oldr != t.root {
			// 建立{后缀链接} [后缀链接是为了新建后缀节点后，快速找到新的活动节点]
			// 详情：oldr和r两个为for循环内分裂产生的后缀节点，上一次分裂产生的节点 指向 新分裂产生的节点
			oldr.linkState = r
		}
		oldr = r

		// 变更{活动节点}
		s, start = t.canonize(s.linkState, start, end-1)
	}
	if oldr != t.root {
		oldr.linkState = r
	}

	// update active point
	t.s = s
	t.start = start
}

// testAndSplit tests whether a state with canonical ref. pair
// (s, (start, end)) is the end point, that is, a state that have
// a c-transition. If not, then state (exs, (start, end)) is made
// explicit (if not already so).
func (t *STree) testAndSplit(s *state, start, end Pos) (exs *state, endPoint bool) {
	c := t.data[t.end] // 新增的节点
	if start <= end {  // 使用end-start表示reminder>0，有未创建的后缀
		tr := s.findTran(t.data[start])
		splitPoint := tr.start + end - start + 1
		if t.data[splitPoint].Val() == c.Val() {
			return s, true
		}
		// 分裂已有的tran
		// make the (s, (start, end)) state explicit
		newSt := newState(s.tree)
		newSt.addTran(splitPoint, tr.end, tr.state)
		tr.end = splitPoint - 1
		tr.state = newSt
		return newSt, false
	}
	if s == t.auxState || s.findTran(c) != nil {
		return s, true
	}
	return s, false
}

// canonize returns updated state and start position for ref. pair
// (s, (start, end)) of state r so the new ref. pair is canonical,
// that is, referenced from the closest explicit ancestor of r.
func (t *STree) canonize(s *state, start, end Pos) (*state, Pos) {
	if s == t.auxState {
		s, start = t.root, start+1
	}
	if start > end { // reminder为0，即不需要增加新的后缀节点
		return s, start
	}

	var tr *tran
	for {
		// 找到活动节点的对应后缀的边
		if start <= end {
			tr = s.findTran(t.data[start])
			if tr == nil {
				panic(fmt.Sprintf("there should be some transition for '%d' at %d",
					t.data[start].Val(), start))
			}
		}
		// 如果该边长度大于新增前缀，则结束
		if tr.end-tr.start > end-start {
			break
		}
		// 如果该边长度小于新增的前缀，则继续向下找
		start += tr.end - tr.start + 1
		s = tr.state
	}
	if s == nil {
		panic("there should always be some suffix link resolution")
	}
	// 返回新的活动节点位置
	return s, start
}

func (t *STree) At(p Pos) Token {
	if p < 0 || p >= Pos(len(t.data)) {
		panic("position out of bounds")
	}
	return t.data[p]
}

func (t *STree) String() string {
	buf := new(bytes.Buffer)
	printState(buf, t.root, 0)
	return buf.String()
}

func printState(buf *bytes.Buffer, s *state, ident int) {
	for _, tr := range s.trans {
		fmt.Fprint(buf, strings.Repeat("  ", ident))
		fmt.Fprintf(buf, "* (%d, %d)\n", tr.start, tr.ActEnd())
		printState(buf, tr.state, ident+1)
	}
}

// state is an explicit state of the suffix tree.
// 后缀树的节点
type state struct {
	tree      *STree  // 所在的树
	trans     []*tran // 扇出的线
	linkState *state  // 后缀链接
}

func newState(t *STree) *state {
	return &state{
		tree:      t,
		trans:     make([]*tran, 0),
		linkState: nil,
	}
}

func (s *state) addTran(start, end Pos, r *state) {
	s.trans = append(s.trans, newTran(start, end, r))
}

// fork creates a new branch from the state s.
func (s *state) fork(i Pos) *state {
	r := newState(s.tree)
	s.addTran(i, infinity, r)
	return r
}

// findTran finds c-transition.
func (s *state) findTran(c Token) *tran {
	for _, tran := range s.trans {
		if s.tree.data[tran.start].Val() == c.Val() {
			return tran
		}
	}
	return nil
}

// tran represents a state's transition.
// 后缀树的边
type tran struct {
	start, end Pos    // 对应{STree.data}的起始结束坐标
	state      *state // 指的是：end的节点/state
}

func newTran(start, end Pos, s *state) *tran {
	return &tran{start, end, s}
}

func (t *tran) len() int {
	return int(t.end - t.start + 1)
}

// ActEnd returns actual end position as consistent with
// the actual length of the data in the STree.
func (t *tran) ActEnd() Pos {
	if t.end == infinity {
		return Pos(len(t.state.tree.data)) - 1
	}
	return t.end
}
