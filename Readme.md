# WSV

This is an implementation of WSV in Go as described by [https://github.com/Stenway/WSV-TS](https://github.com/Stenway/WSV-TS).

## Getting Started

```bash
go get https://github.com/internetcalifornia/wsv
```

## Reading Usage

When using the reader you can read line by line, `Read()` or use the convenient `ReadAll()` function which reads all lines into a slice of rows of records/fields.

```go
package main

import (
    "strings"
    "github.com/internetcalifornia/wsv"
)

func main() {
    lines := make([]string, 0)
    lines = append(lines, `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`)
    lines = append(lines, `"Jean Smith" "Le Croix" "Jan 01 2023" "Space Purple"`)
    lines = append(lines, `"Mary Jane" "Vasquez Rojas" "Feb 02 2021" "Midnight Grey"`)
    file := strings.Join(lines, string('\n'))
    str := strings.NewReader(file)
    r := wsv.NewReader(str)

    // read line by line
    line := r.Read()

    // or read until end of file
    // records := r.ReadAll()
}
```
