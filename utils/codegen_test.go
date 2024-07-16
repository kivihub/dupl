package utils

import (
	"testing"
)

func TestIsCodegen(t *testing.T) {
	testCases := []struct {
		filePath string
		expect   bool
	}{
		{"../_input_example/demo.txt", false},
		{"../_input_example/codegen.txt", true},
	}
	for _, tc := range testCases {
		actual := IsCodegen(tc.filePath)
		if tc.expect != actual {
			t.Errorf("%s got '%t', want '%t'", tc.filePath, actual, tc.expect)
		}
	}
}
