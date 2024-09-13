package main

import (
	"testing"
)

func TestIsCodegen(t *testing.T) {
	v := true
	verbose = &v
	ignoreFileContentExpr = "// Code generated"
	InitIgnoreRegexBeforeAnalyze()

	testCases := []struct {
		filePath string
		expect   bool
	}{
		{"_input_example/codegen.txt", true},
	}
	for _, tc := range testCases {
		actual := IgnoreFile(tc.filePath)
		if tc.expect != actual {
			t.Errorf("%s got '%t', want '%t'", tc.filePath, actual, tc.expect)
		}
	}
}
