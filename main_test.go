package main

import (
	"bytes"
	"fmt"
	"github.com/bytedance/mockey"
	"github.com/kivihub/dupl/context"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/goconvey/convey"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDuplForPath(t *testing.T) {
	context.IsDebug = true
	filePath := []string{"_input_example/clone_left.txt", "_input_example/clone_right.txt"}
	filePath = InsertPackageInfo(filePath)
	os.Args = []string{"dupl", "-t=100", "-ft=20", "-fr=30", "-ignoreCodegen", "-plumbing", "-verbose"}
	//os.Args = []string{"dupl", "-t=100", "-plumbing", "-verbose"}
	runMockMain(t, filePath, func(output string) {
		convey.So(strings.Count(output, "duplicate of"), assertions.ShouldEqual, 1)
	})
}

func TestDuplForDir(t *testing.T) {
	dir := "."
	os.Args = []string{"dupl", "-t=100", "-ft=20", "-fr=30", "-ignoreCodegen", "-maxFileSize=1048576", "-plumbing", "-verbose", dir}
	main()
}

func InsertPackageInfo(filePaths []string) []string {
	pkg := []byte("package demo\n")
	ret := make([]string, len(filePaths))
	for i, path := range filePaths {
		dir, file := filepath.Split(path)
		newPath := dir + "." + file
		bytes, _ := os.ReadFile(path)
		bytes = append(pkg, bytes...)
		os.WriteFile(newPath, bytes, os.FileMode(0666))
		ret[i] = newPath
	}
	return ret
}

func runMockMain(t *testing.T, filePath []string, processOutput func(string)) {
	mockey.PatchConvey("TestMockPath", t, func() {
		mockey.Mock(filesFeed).Return(func() chan string {
			fchan := make(chan string)
			go func() {
				for _, file := range filePath {
					fchan <- file
				}
				close(fchan)
			}()
			return fchan
		}()).Build()
		endCapture := CaptureStdout()
		main()
		capturedContent := endCapture()
		fmt.Print("Captured: ", capturedContent)
		processOutput(capturedContent)
	})
}

// CaptureStdout 目前只适合拦截少量输出，如果过大超过缓冲区，则会阻塞
func CaptureStdout() func() string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	return func() string {
		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		os.Stdout = oldStdout
		return buf.String()
	}
}
