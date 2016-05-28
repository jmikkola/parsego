 [![Build Status](https://travis-ci.org/jmikkola/parsego.svg?branch=master)](https://travis-ci.org/jmikkola/parsego) [![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/jmikkola/parsego/parser)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmikkola/parsego)](https://goreportcard.com/report/github.com/jmikkola/parsego) [![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/jmikkola/parsego/blob/fix-badges/LICENSE)

# parsego

A simple, easy to use parser-combinator written in Golang.

Example usage:

```go
package main

import "fmt"
import "github.com/jmikkola/parsego/parser"

func main() {
    p := parser.Sequence(
        parser.Digits(),
        parser.Maybe(
            parser.Sequence(
                parser.Char('.'),
                parser.Digits())))
    result, err := parser.ParseString(p, "1234.567")
    if err != nil {
        fmt.Println("failed to parse", err)
    } else {
        fmt.Println("parsed", result)
    }
}
```

See examples/ for more examples.
