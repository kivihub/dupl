# dupl [![Build Status](https://travis-ci.org/mibk/dupl.png)](https://travis-ci.org/mibk/dupl)

**dupl** is a tool written in Go for finding code clones. So far it can find clones only
in the Go source files. The method uses suffix tree for serialized ASTs. It ignores values
of AST nodes. It just operates with their types (e.g. `if a == 13 {}` and `if x == 100 {}` are
considered the same provided it exceeds the minimal token sequence size).

Due to the used method dupl can report so called "false positives" on the output. These are
the ones we do not consider clones (whether they are too small, or the values of the matched
tokens are completely different).

## Installation

```bash
go get -u github.com/kivihub/dupl
```

## Usage

```
Usage of dupl:
  dupl [flags] [paths]

Paths:
  If the given path is a file, dupl will use it regardless of
  the file extension. If it is a directory it will recursively
  search for *.go files in that directory.

  If no path is given dupl will recursively search for *.go
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
        explain what is being done
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
  -ignorePath     
        ignore files it's path matching the given regexp
  -ignoreContent  
        ignore files it's content matching the given regexp
  -maxFileSize
        ignore file when file size exceed maxFileSize

Examples:
  dupl -t 100
        Search clones in the current directory of size at least
        100 tokens.
  dupl $(find app/ -name '*_test.go')
        Search for clones in tests in the app directory.
  find app/ -name '*_test.go' |dupl -files
        The same as above.
```

## Example

The reduced output of this command with the following parameters for the [Docker](https://www.docker.com) source code
looks like [this](http://htmlpreview.github.io/?https://github.com/kivihub/dupl/blob/master/_output_example/docker.html).

```bash
$ dupl -t 200 -html >docker.html
```
