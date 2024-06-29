package document

import (
	"errors"
	"fmt"

	"github.com/internetcalifornia/wsv/v2/record"
)

var (
	ErrNotEnoughLines = errors.New("document does not have more than 1 line")
)

type documentLine struct {
	doc          *Document
	fields       []record.RecordField
	comment      string
	currentField int
	// Lines are 1-indexed
	line int
	// count of data fields, has a getter documentLine.FieldCount()
	fieldCount int
}

type DocumentLine interface {
	// determine if tabular document line is valid based on the number of lines of the first row/header, returns true, nil if has the correct number of data fields
	//
	// returns false, and an error documenting the difference
	Validate() (bool, error)
	// Append a value to the end of the line
	Append(val string) error
	// Append a null value to the end of the line
	AppendNull() error
	// Get the next field value, or error if at the end of the line for data
	NextField() (*record.RecordField, error)
	// Record at index, or Field Not found, field 0-indexed
	Field(fieldIndex int) (*record.RecordField, error)
	// A count of the number of data fields in the line
	FieldCount() int
	// Get the line number
	LineNumber() int
	// Update the value of a particular field
	UpdateField(fieldIndex int, val string) error
	// Update a comment on the line
	UpdateComment(val string)
	// Get the value of comment for the line
	Comment() string
	// Update the field name for the field at the given index
	//
	// ErrFieldNotFound is returned if there is no field at the
	UpdateFieldName(fieldIndex int, val string) error
}

func (line *documentLine) Validate() (bool, error) {
	if !line.doc.Tabular {
		return true, nil
	}
	if line.doc.HasHeaders() && len(line.doc.Headers()) != line.fieldCount {
		return false, fmt.Errorf("line %d does not have the correct number of fields %d/%d (current/expected)", line.line, line.fieldCount, len(line.doc.Headers()))
	}
	line1, err := line.doc.Line(1)
	if err != nil {
		return false, err
	}
	if !line.doc.HasHeaders() && line.line > 1 && line1.FieldCount() >= 1 && line1.FieldCount() != line.fieldCount {
		return false, fmt.Errorf("line %d does not have the correct number of fields %d/%d (current/expected)", line.line, line.FieldCount(), line1.FieldCount())
	}
	return true, nil
}

func (line *documentLine) Append(val string) error {
	field := record.RecordField{
		Value: val,
	}
	if line.doc.HasHeaders() && (line.doc.headerLine == 0 || line.line == line.doc.headerLine) {
		field.IsHeader = true
		field.FieldName = val
		line.doc.headerLine = line.line
	}
	fieldInd := len(line.fields)
	err := line.checkFieldIndex(fieldInd)
	if err != nil {
		return ErrFieldCount
	}

	if line.line > line.doc.headerLine && len(line.doc.Headers())-1 > fieldInd {
		field.FieldName = line.doc.Headers()[fieldInd]
	}
	field.FieldIndex = fieldInd
	line.fields = append(line.fields, field)
	fw := field.CalculateFieldLength()
	line.doc.SetMaxColumnWidth(fieldInd, fw)
	if line.doc.HasHeaders() && line.line == line.doc.headerLine {
		line.doc.AppendHeader(val)
	}
	// increment the field count for the line
	line.fieldCount++
	return nil
}

func (line *documentLine) AppendNull() error {
	field := record.RecordField{IsNull: true}
	if line.doc.HasHeaders() && (line.doc.headerLine == 0 || line.line == line.doc.headerLine) {
		field.IsHeader = true
		field.FieldName = "-"
		line.doc.headerLine = line.line
	}
	fieldInd := len(line.fields)
	err := line.checkFieldIndex(fieldInd)
	if err != nil {
		return ErrFieldCount
	}

	if line.line > line.doc.headerLine && len(line.doc.Headers())-1 > fieldInd {
		field.FieldName = line.doc.Headers()[fieldInd]
	}
	field.FieldIndex = fieldInd
	line.fields = append(line.fields, field)
	fw := field.CalculateFieldLength()
	line.doc.SetMaxColumnWidth(fieldInd, fw)
	if line.doc.HasHeaders() && line.line == line.doc.headerLine {
		line.doc.AppendHeader("-")
	}
	// increment the field count for the line
	line.fieldCount++
	return nil
}

func (line *documentLine) NextField() (*record.RecordField, error) {
	if len(line.fields)-1 < line.currentField {
		return nil, ErrFieldNotFound
	}
	fieldInd := line.currentField
	line.currentField++
	return &line.fields[fieldInd], nil
}

// Returns the number of data fields, non-comment fields
func (line *documentLine) FieldCount() int {
	return line.fieldCount
}

func (line *documentLine) LineNumber() int {
	return line.line
}

// Update field at line `fi`, `fi` is 0-index
func (line *documentLine) UpdateField(fieldInd int, val string) error {
	if len(line.fields)-1 < fieldInd || fieldInd < 0 {
		return ErrFieldNotFound
	}
	field := line.fields[fieldInd]
	field.Value = val
	line.fields[fieldInd] = field
	fw := field.CalculateFieldLength()
	line.doc.SetMaxColumnWidth(fieldInd, fw)
	return nil
}

func (line *documentLine) UpdateComment(val string) {
	line.comment = val
}

func (line *documentLine) Comment() string {
	return line.comment
}

// check field index is valid, returns the number of fields left, -1 is returned when document is not tabular or is the first line
func (line *documentLine) checkFieldIndex(fieldInd int) error {
	if !line.doc.Tabular || line.line == line.doc.headerLine {
		// if the document is not tabular or this is the first line this check won't be in effect
		return nil
	}
	if line.doc.LineCount() <= 1 {
		// unexpected
		return ErrNotEnoughLines
	}
	headerLine, err := line.doc.Line(line.doc.headerLine)
	if err != nil {
		return ErrLineNotFound
	}

	fc := headerLine.FieldCount() - fieldInd

	if fc <= 0 {
		return &WriteError{err: ErrFieldCount, line: line.line, fieldIndex: fieldInd}
	}
	return nil
}

func (line *documentLine) UpdateFieldName(fi int, val string) error {
	if len(line.fields)-1 < fi {
		return ErrFieldCount
	}
	field := line.fields[fi]
	field.FieldName = val
	line.fields[fi] = field
	return nil
}

func (line *documentLine) Field(fieldIndex int) (*record.RecordField, error) {
	if len(line.fields)-1 < fieldIndex {
		return nil, ErrFieldNotFound
	}
	return &line.fields[fieldIndex], nil
}
