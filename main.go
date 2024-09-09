package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kivihub/dupl/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kivihub/dupl/job"
	"github.com/kivihub/dupl/printer"
	"github.com/kivihub/dupl/syntax"
)

const defaultThreshold = 15

var (
	paths     = []string{"."}
	vendor    = flag.Bool("vendor", false, "")
	verbose   = flag.Bool("verbose", false, "")
	threshold = flag.Int("threshold", defaultThreshold, "") // Token阈值
	files     = flag.Bool("files", false, "")
	html      = flag.Bool("html", false, "")
	plumbing  = flag.Bool("plumbing", false, "")

	funcThreshold = flag.Int("funcThreshold", 0, "") // 函数重复行数阈值
	funcRatio     = flag.Int("funcRatio", 0, "")     // 重复行数占所在函数总行数的最小比例阈值：[-100, 100], 与-plumbing配合使用
	ignoreCodegen = flag.Bool("ignoreCodegen", false, "")
	maxFileSize   = flag.String("maxFileSize", "", "") // 支持的最大文件大小，防止大文件阻塞
)

const (
	vendorDirPrefix = "vendor" + string(filepath.Separator)
	vendorDirInPath = string(filepath.Separator) + vendorDirPrefix
)

func init() {
	flag.BoolVar(verbose, "v", false, "alias for -verbose")
	flag.IntVar(threshold, "t", defaultThreshold, "alias for -threshold")
	flag.IntVar(funcThreshold, "ft", 0, "alias for -funcThreshold")
	flag.IntVar(funcRatio, "fr", 0, "alias for -funcRatio")
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if *html && *plumbing {
		log.Fatal("you can have either plumbing or HTML output")
	}
	if flag.NArg() > 0 {
		paths = flag.Args()
	}

	if *verbose {
		log.Println("Building suffix tree")
	}
	syntax.InitFuncDuplManager(*funcRatio, *verbose)
	maxFileSizeInt := utils.ParseStorageToBytes(*maxFileSize)
	log.Printf("Flag maxFileSize: %s -> %d Bytes", *maxFileSize, maxFileSizeInt)
	schan := job.Parse(filesFeed(), int(maxFileSizeInt))
	t, data, done := job.BuildTree(schan)
	<-done

	// finish stream
	// 后缀树增加{结束节点$}
	t.Update(&syntax.Node{Type: -1})

	if *verbose {
		log.Println("Searching for clones")
	}
	mchan := t.FindDuplOver(*threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			var match syntax.Match
			if *funcThreshold == 0 {
				match = syntax.FindSyntaxUnits(*data, m, *threshold)
			} else {
				match = syntax.FindFuncUnits(*data, m, *funcThreshold, *verbose)
			}
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	newPrinter := printer.NewText
	if *html {
		newPrinter = printer.NewHTML
	} else if *plumbing {
		newPrinter = printer.NewPlumbing
	}
	p := newPrinter(os.Stdout, ioutil.ReadFile)
	if err := printDupls(p, duplChan); err != nil {
		log.Fatal(err)
	}
}

func filesFeed() chan string {
	if *files {
		fchan := make(chan string)
		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				fchan <- strings.TrimPrefix(f, "./")
			}
			close(fchan)
		}()
		return fchan
	}
	return crawlPaths(paths)
}

func crawlPaths(paths []string) chan string {
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				log.Fatal(err)
			}
			if !info.IsDir() {
				fchan <- path
				continue
			}
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if !*vendor && (strings.HasPrefix(path, vendorDirPrefix) ||
					strings.Contains(path, vendorDirInPath)) {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					if *ignoreCodegen && utils.IsCodegen(path) {
						return nil
					}
					fchan <- path
				}
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}
		close(fchan)
	}()
	return fchan
}

// 通过两个逻辑保留最长重复串：
// 1）由于是深度遍历所以{duplChan}会首先读取到长串，然后append到groups
// 2）通过[filename & Position]进行Uniq去重，去重后面出现的较短的重复串
func printDupls(p printer.Printer, duplChan <-chan syntax.Match) error {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := p.PrintHeader(); err != nil {
		return err
	}
	syntax.GlobalFuncDuplManager.RemoveFuncLessRatio()
	for _, k := range keys {
		uniq := unique(groups[k])
		if len(uniq) > 1 {
			// 打印本重复组
			if err := p.PrintClones(uniq); err != nil {
				return err
			}
		}
	}
	return p.PrintFooter()
}

// 按 [Filename & Position] 去重
func unique(group [][]*syntax.Node) [][]*syntax.Node {
	fileMap := make(map[string]map[int]struct{})

	var newGroup [][]*syntax.Node
	for _, seq := range group {
		node := seq[0]
		file, ok := fileMap[node.Filename]
		if !ok {
			file = make(map[int]struct{})
			fileMap[node.Filename] = file
		}
		if _, ok := file[node.Pos]; !ok {
			file[node.Pos] = struct{}{}
			newGroup = append(newGroup, seq)
		}
	}
	return newGroup
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: dupl [flags] [paths]

Paths:
  If the given path is a file, dupl will use it regardless of
  the file extension. If it is a directory, it will recursively
  search for *.go files in that directory.

  If no path is given, dupl will recursively search for *.go
  files in the current directory.

Flags:
  -files
    	read file names from stdin one at each line
  -html
    	output the results as HTML, including duplicate code fragments
  -plumbing
    	plumbing (easy-to-parse) output for consumption by scripts or tools
  -t, -threshold size
    	minimum token sequence size as a clone (default 15)
  -vendor
    	check files in vendor directory
  -v, -verbose
    	explain what is being do
  -ft, -funcThreshold num
     	minimum lines of function duplicate, plumbing output change to 
     	function's duplicates instead of file's duplicate
  -fr, -funcRatio num
        minimum proportion of duplicate lines to the total number of lines 
        within its function, value range is [-100, 100]. 
        A positive number means that both duplicate function pairs need to 
        exceed the threshold, and a negative number means that any function 
        of the duplicate function pair can exceed the threshold. 
        used with flag -plumbing
  -ignoreCodegen
        ignore codegen file, accelerate parsing speed
  -maxFileSize
        ignore file when file size exceed maxFileSize

Examples:
  dupl -t 100
    	Search clones in the current directory of size at least
    	100 tokens.
  dupl $(find app/ -name '*_test.go')
    	Search for clones in tests in the app directory.
  find app/ -name '*_test.go' |dupl -files
    	The same as above.`)
	os.Exit(2)
}
