package job

import (
	"log"

	"github.com/kivihub/dupl/syntax"
	"github.com/kivihub/dupl/syntax/golang"
)

func Parse(fchan chan string) chan []*syntax.Node {

	// parse AST
	achan := make(chan *syntax.Node)
	go func() {
		for file := range fchan {
			ast, err := golang.Parse(file)
			if err != nil {
				log.Println(err)
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
