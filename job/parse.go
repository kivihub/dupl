package job

import (
	"log"
	"os"

	"github.com/kivihub/dupl/syntax"
	"github.com/kivihub/dupl/syntax/golang"
)

func Parse(fchan chan string, maxFileSize int) chan []*syntax.Node {
	// parse AST
	achan := make(chan *syntax.Node)
	go func() {
		for file := range fchan {
			if fileInfo, err := os.Stat(file); err != nil {
				log.Printf("Parsing %s\n", file)
			} else {
				fileSize := int(fileInfo.Size())
				if fileSize > maxFileSize {
					log.Printf("Ignore Parsing %s %dbytes excced threshold %d\n", file, fileSize, maxFileSize)
					continue
				} else {
					log.Printf("Parsing %s size:%d bytes\n", file, fileSize)
				}
			}

			ast, err := golang.Parse(file)
			if err != nil {
				log.Println("Parse error", err)
				continue
			}
			achan <- ast
		}
		close(achan)
	}()

	// serialize
	schan := make(chan []*syntax.Node)
	go func() {
		for ast := range achan {
			// seq是一个.go文件的所有SyntaxNode扁平集合
			seq := syntax.Serialize(ast)
			schan <- seq
		}
		close(schan)
	}()
	return schan
}
