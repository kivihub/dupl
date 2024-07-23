package suffixtree

import "sort"

type Match struct {
	Ps  []Pos
	Len Pos
}

type posList struct {
	positions []Pos
}

func newPosList() *posList {
	return &posList{make([]Pos, 0)}
}

func (p *posList) append(p2 *posList) {
	p.positions = append(p.positions, p2.positions...)
}

func (p *posList) add(pos Pos) {
	p.positions = append(p.positions, pos)
}

type contextList struct {
	lists map[int]*posList
}

func newContextList() *contextList {
	return &contextList{make(map[int]*posList)}
}

func (c *contextList) getAll() []Pos {
	keys := make([]int, 0, len(c.lists))
	for k := range c.lists {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var ps []Pos
	for _, k := range keys {
		ps = append(ps, c.lists[k].positions...)
	}
	return ps
}

func (c *contextList) append(c2 *contextList) {
	for lc, pl := range c2.lists {
		if _, ok := c.lists[lc]; ok {
			c.lists[lc].append(pl)
		} else {
			c.lists[lc] = pl
		}
	}
}

// FindDuplOver find pairs of maximal duplicities over a threshold
// length.
func (t *STree) FindDuplOver(threshold int) <-chan Match {
	auxTran := newTran(0, 0, t.root)
	ch := make(chan Match)
	go func() {
		walkTrans(auxTran, 0, threshold, ch)
		close(ch)
	}()
	return ch
}

// DPS遍历后缀树
// length为根结点走完parent的长度，会重复返回，即返回起始点相同但长度不同的Match结果，在print时进行过滤
func walkTrans(parent *tran, length, threshold int, ch chan<- Match) *contextList {
	s := parent.state

	// 记录后缀树从根结点到叶子结点路径的起始Pos(在STree.data的位置)
	cl := newContextList()

	// 叶子结点
	if len(s.trans) == 0 {
		pl := newPosList()
		start := parent.end + 1 - Pos(length) // start为重复的起始Pos
		pl.add(start)
		ch := 0
		if start > 0 {
			ch = s.tree.data[start-1].Val() // ch为重复的起始Pos对应的字符/值
		}
		cl.lists[ch] = pl // ch和start 两者结合length就可以获取重复的内容
		return cl
	}

	// 非叶子结点：递归遍历
	for _, t := range s.trans {
		ln := length + t.len()
		cl2 := walkTrans(t, ln, threshold, ch)
		if ln >= threshold { // 感觉不加也可以，因为下面有length阈值判断
			cl.append(cl2)
		}
	}
	// 非叶子结点，判断超过threshold则，把{STree.data}的起始Pos结合和长度返回
	if length >= threshold && len(cl.lists) > 1 {
		m := Match{cl.getAll(), Pos(length)}
		ch <- m
	}
	return cl
}
