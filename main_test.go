package main

import (
	"bytes"
	"fmt"
	"github.com/bytedance/mockey"
	"github.com/kivihub/dupl/context"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
	"testing"
)

func TestNewFlag(t *testing.T) {
	context.IsDebug = true
	filePath := []string{"_input_example/clone_left.txt", "_input_example/clone_right.txt"}
	os.Args = []string{"dupl", "-t=100", "-ft=20", "-fr=30", "-ignoreCodegen", "-plumbing", "-verbose"}
	//os.Args = []string{"dupl", "-t=100", "-plumbing", "-verbose"}
	runMockMain(t, filePath, func(output string) {
		convey.So(strings.Count(output, "duplicate of"), assertions.ShouldEqual, 1)
	})
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
