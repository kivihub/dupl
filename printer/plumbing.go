package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kivihub/dupl/syntax"
)

type plumbing struct {
	printedDuplPairs map[string]bool
	w                io.Writer
	ReadFile
}

func NewPlumbing(w io.Writer, fread ReadFile) Printer {
	return &plumbing{make(map[string]bool), w, fread}
}

func (p *plumbing) PrintHeader() error { return nil }

func (p *plumbing) PrintClones(dups [][]*syntax.Node) error {
	clones, err := prepareClonesInfo(p.ReadFile, dups)
	if err != nil {
		return err
	}
	sort.Sort(byNameAndLine(clones))
	for i, cl := range clones {
		nextCl := clones[(i+1)%len(clones)]
		if p.hasPrinted(cl.filename, cl.lineStart, cl.lineEnd, nextCl.filename, nextCl.lineStart, nextCl.lineEnd) {
			continue
		}

		f1ExceedRatio := syntax.GlobalFuncDuplManager.Exist(cl.filename, cl.lineStart)
		f2ExceedRatio := syntax.GlobalFuncDuplManager.Exist(nextCl.filename, nextCl.lineStart)

		var printDuplPair bool
		if syntax.GlobalFuncDuplManager.BothFuncNeedExceedRatio() {
			printDuplPair = f1ExceedRatio && f2ExceedRatio
		} else {
			printDuplPair = f1ExceedRatio || f2ExceedRatio
		}
		if printDuplPair {
			fmt.Fprintf(p.w, "%s:%d-%d: duplicate of %s:%d-%d\n", cl.filename, cl.lineStart, cl.lineEnd, nextCl.filename, nextCl.lineStart, nextCl.lineEnd)
		}

	}
	return nil
}

func (p *plumbing) PrintFooter() error { return nil }

func (p *plumbing) hasPrinted(f1 string, st1, end1 int, f2 string, st2, end2 int) bool {
	left := fmt.Sprintf("%s:%d-%d", f1, st1, end1)
	right := fmt.Sprintf("%s:%d-%d", f2, st2, end2)

	duplPair := []string{left, right}
	sort.Strings(duplPair)

	key := strings.Join(duplPair, "#")
	if _, exist := p.printedDuplPairs[key]; exist {
		return true
	} else {
		p.printedDuplPairs[key] = true
		return false
	}
}
