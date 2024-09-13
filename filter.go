package main

import (
	"log"
	"os"
	"regexp"
)

var ignoredFiles = make(map[string]bool)
var notIgnoredFiles = make(map[string]bool)

func IgnoreFile(filePath string) bool {
	if !setIgnoreRegexp {
		return false
	}

	if _, exist := notIgnoredFiles[filePath]; exist {
		return false
	}
	if _, exist := ignoredFiles[filePath]; exist {
		return true
	}

	if IgnoreFileByPath(filePath) || IgnoreFileByContent(filePath) {
		ignoredFiles[filePath] = true
		return true
	} else {
		notIgnoredFiles[filePath] = true
		return false
	}
}

var ignoreFileContentRegexp *regexp.Regexp
var ignoreFilePathRegexp *regexp.Regexp
var setIgnoreRegexp bool

func InitIgnoreRegexBeforeAnalyze() {
	if *verbose {
		log.Printf("--Ignore file path regex is: %s\n", defaultVal(*ignoreFilePathExpr, "<not_set>"))
		log.Printf("--Ignore file content regex is: %s\n", defaultVal(*ignoreFileContentExpr, "<not_set>"))
	}
	ignoreFilePathRegexp = prepareRegexp(*ignoreFilePathExpr)
	ignoreFileContentRegexp = prepareRegexp(*ignoreFileContentExpr)
	if ignoreFileContentRegexp != nil || ignoreFilePathRegexp != nil {
		setIgnoreRegexp = true
	}
}

func IgnoreFileByPath(filePath string) bool {
	ignore := ignoreFilePathRegexp != nil && ignoreFilePathRegexp.MatchString(filePath)
	if *verbose && ignore {
		log.Printf("IgnoreFileByPath: %s\n", filePath)
	}
	return ignore
}

func IgnoreFileByContent(filePath string) bool {
	if ignoreFileContentRegexp == nil {
		return false
	}

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文件时出错:%s, %v\n", filePath, err)
		return false
	}

	ignore := ignoreFileContentRegexp.MatchString(string(contentBytes))
	if *verbose && ignore {
		log.Printf("IgnoreFileByContent: %s\n", filePath)
	}
	return ignore
}

func defaultVal(val, def string) string {
	if val != "" {
		return val
	}
	return def
}

func prepareRegexp(expr string) *regexp.Regexp {
	if expr == "" {
		return nil
	}

	ret, err := regexp.Compile(expr)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}
