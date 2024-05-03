package main

import (
	"fmt"

	wsv "github.com/internetcalifornia/wsv/v2/document"
)

func main() {
	doc := wsv.NewDocument()

	line, err := doc.AddLine()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = line.Append("name")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = line.Append("age")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = line.Append("favorite color")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = doc.AppendLine(wsv.Field("scott"), wsv.Null(), wsv.Field("red"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
