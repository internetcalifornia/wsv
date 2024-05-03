package main_test

import (
	"fmt"
	"os"
	"testing"

	wsv "github.com/internetcalifornia/wsv/v2/reader"
)

func TestRead(t *testing.T) {
	dir, ok := os.LookupEnv("PROJECT_DIR")
	if !ok {
		t.Error("PROJECT_DIR env not FOUND")
		t.FailNow()
		return
	}
	file, err := os.Open(fmt.Sprintf("%s/examples/sample.wsv", dir))
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	r := wsv.NewReader(file)
	lines, err := r.ReadAll()
	if err != nil {
		t.Error(err)
		return
	}
lineLoop:
	for _, line := range lines {
		for {
			// field
			_, err := line.NextField()
			if err == wsv.ErrEndOfLine {
				continue lineLoop
			}
			// field.SerializeText()
			// field.Value
			// field.FieldName
		}
	}
}
