# parsego

[![Build Status](https://travis-ci.org/jmikkola/parsego.svg?branch=master)](https://travis-ci.org/jmikkola/parsego)

A poorly-designed parser combinator in Go.

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

See also `parser.ParseScanner()` for parsing larger input.
