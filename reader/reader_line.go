package reader

import (
	"github.com/internetcalifornia/wsv/v1/internal"
)

type ReaderLine interface {
	Field(fi int) (*internal.RecordField, error)
	// Get the value of comment for the line
	Comment() string
	// Get the line number
	LineNumber() int
	// A count of the number of data fields in the line
	FieldCount() int
	// Get the next field value, or error if at the end of the line for data
	NextField() (*internal.RecordField, error)
}

type readerLine struct {
	fields  []internal.RecordField
	comment string
	// Lines are 1-indexed
	line int
	// count of data fields, has a getter readerLine.FieldCount()
	fieldCount   int
	currentField int
}

func (line *readerLine) NextField() (*internal.RecordField, error) {
	if len(line.fields)-1 < line.currentField {
		return nil, internal.ErrFieldNotFound
	}
	fieldInd := line.currentField
	line.currentField++
	return &line.fields[fieldInd], nil
}

// Returns the number of data fields, non-comment fields
func (line *readerLine) FieldCount() int {
	return line.fieldCount
}

func (line *readerLine) LineNumber() int {
	return line.line
}

func (line *readerLine) Field(fieldIndex int) (*internal.RecordField, error) {
	if len(line.fields)-1 < fieldIndex {
		return nil, internal.ErrFieldNotFound
	}
	return &line.fields[fieldIndex], nil
}

func (line *readerLine) Comment() string {
	return line.comment
}

func (line *readerLine) UpdateComment(val string) {
	line.comment = val
}
