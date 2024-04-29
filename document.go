package wsv

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	ErrLineNotFound       = errors.New("line does not exist")
	ErrFieldNotFound      = errors.New("field does not exist")
	ErrOmitHeaders        = errors.New("document configured to omit headers")
	ErrInvalidPaddingRune = errors.New("only whitespace characters can be used for padding")
	ErrStartedToWrite     = errors.New("document started to write, need to reset document to edit")
	ErrLineIsNotHeader    = errors.New("the line is not the first line in the document")
	ErrNotEnoughLines     = errors.New("document does not have more than 1 line")
	ErrLineNotEditable    = errors.New("cannot edit")
)

type WriteError struct {
	line        int
	fieldIndex  int
	headerCount int

	err error
}

func (e *WriteError) Error() string {

	if e.err == ErrFieldCount {
		return fmt.Sprintf("trying to write a value on line %d for the field index %d is not allowed because the document has %d header(s) and the writer is configured to be tabular.", e.line, e.fieldIndex, e.headerCount)
	}

	if e.err == ErrStartedToWrite {
		return fmt.Sprintf("the writer already started and has currently written up until line %d, reset the writer to continue editing the document.", e.line)
	}

	if e.err == ErrLineNotEditable {
		return fmt.Sprintf("line %d is non-editable %s", e.line, e.err.Error())
	}

	if e.err == ErrFieldNotFound {
		return fmt.Sprintf("field %d, %s for line %d", e.fieldIndex, e.err.Error(), e.line)
	}

	return e.err.Error()
}

type document struct {
	Tabular          bool
	EmitHeaders      bool
	lines            []*documentLine
	MaxColumnWidth   map[int]int
	padding          []rune
	currentWriteLine int
	currentField     int
	startedWriting   bool
	Headers          []string
	HasHeaders       bool
}

func NewDocument() *document {
	doc := document{
		Tabular:          true,
		EmitHeaders:      true,
		lines:            make([]*documentLine, 0),
		currentWriteLine: 0,
		currentField:     0,
		MaxColumnWidth:   make(map[int]int, 0),
		startedWriting:   false,
		// The runes in between data values
		padding:    []rune{' ', ' '},
		Headers:    make([]string, 0),
		HasHeaders: true,
	}
	return &doc
}

func (doc *document) SetPadding(rs []rune) error {
	for _, r := range rs {
		if !isFieldDelimiter(r) {
			return &WriteError{err: ErrInvalidPaddingRune}
		}
	}
	doc.padding = rs
	return nil
}

type documentLine struct {
	doc     *document
	fields  []RecordField
	Comment string
	// Lines are 1-indexed
	line int
	// count of data fields, has a getter documentLine.FieldCount()
	fieldCount int
}

func (doc *document) AddLine() (*documentLine, error) {
	if doc.startedWriting {
		return nil, &WriteError{err: ErrStartedToWrite, line: doc.currentWriteLine}
	}
	pln := len(doc.lines)
	line := documentLine{
		doc:    doc,
		fields: make([]RecordField, 0),
		line:   pln + 1,
	}

	doc.lines = append(doc.lines, &line)
	doc.currentField = 0
	return &line, nil
}

// Returns the document at the ln specified. Lines are 1-index. If the line does not exist there is an
// ErrLineNotFound error
func (doc *document) Line(ln int) (*documentLine, error) {
	if len(doc.lines)-1 < ln-1 || ln < 1 {
		return nil, ErrLineNotFound
	}
	line := doc.lines[ln-1]
	return line, nil
}

func (doc *document) ResetWrite() {
	doc.startedWriting = false
	doc.currentWriteLine = 0
}

func CalculateFieldLength(f RecordField) int {
	v := f.SerializeText()

	return utf8.RuneCountInString(v)
}

// Write, writes the currently line to a slice of bytes based on the current line in process, calling write will increment the counter after each successful call.
// Once all lines are process will return will return empty slice, EOF
func (doc *document) Write() ([]byte, error) {
	doc.startedWriting = true
	buf := make([]byte, 0)

	if len(doc.lines)-1 < doc.currentWriteLine {
		return buf, io.EOF
	}

	line := doc.lines[doc.currentWriteLine]
	if doc.HasHeaders && !doc.EmitHeaders && doc.currentWriteLine == 0 {
		return buf, ErrOmitHeaders
	}
	// if configured to be tabular, not an empty line, and has too little/many fields compared to headers return an error
	if doc.Tabular && doc.currentWriteLine != 0 && line.fieldCount != 0 && line.fieldCount != len(doc.Headers) {
		return buf, &WriteError{line: line.line, headerCount: len(doc.Headers), fieldIndex: line.fieldCount, err: ErrFieldCount}
	}

	for i, field := range line.fields {
		mw := doc.MaxColumnWidth[i]
		v := field.SerializeText()
		p := utf8.RuneCountInString(v)
		if doc.Tabular && (len(line.fields)-1 != i) {
			for {
				// pad value with single spaces unless it's the last column or line has a comment
				if p < mw {
					v = fmt.Sprintf("%s%s", v, " ")
					p = utf8.RuneCountInString(v)
					continue
				}
				break
			}
		}

		if i == 0 {
			buf = append(buf, []byte(v)...)
		} else {
			buf = append(buf, runeToBytes(doc.padding)...)
			buf = append(buf, []byte(v)...)
		}
	}
	if len(line.Comment) > 0 {
		if len(buf) > 0 {
			buf = append(buf, runeToBytes(doc.padding)...)
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment))...)
		} else {
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment))...)

		}
	}
	buf = append(buf, byte('\n'))
	doc.currentWriteLine += 1
	return buf, nil
}

func (doc *document) WriteAll() ([]byte, error) {
	data := make([]byte, 0)
	for {
		d, err := doc.Write()
		if err == io.EOF {
			break
		}
		if err != nil {

			return data, err
		}
		data = append(data, d...)
	}
	return data, nil
}

func (doc *document) LineCount() int {
	return len(doc.lines)
}

func runeToBytes(rs []rune) []byte {
	b := []byte{}
	for _, r := range rs {
		b = append(b, byte(r))
	}
	return b
}

// Returns a comment if one exists for the rows or an error if comment does not exist
// lines are 1-indexed
func (doc *document) CommentFor(ln int) (string, error) {
	if len(doc.lines) < ln {
		return "", fmt.Errorf("there are no records found for row %d, please ensure you are indexing as 1-indexed values", ln)
	}
	line := doc.lines[ln-1]

	if len(line.Comment) > 0 {
		return line.Comment, nil
	}
	msg := fmt.Errorf("comment not found for row %d", ln)
	return "", msg
}

func (f *RecordField) SerializeText() string {
	wrapped := false
	if f.IsNull {
		return "-"
	}
	v := f.Value

	v = strings.ReplaceAll(v, `"`, `""`)
	if strings.Contains(v, `""`) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if strings.Contains(v, "-") {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	v = strings.ReplaceAll(v, "\n", `"/"`)
	if strings.Contains(v, `"/"`) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if strings.ContainsFunc(v, isFieldDelimiter) && !wrapped {
		wrapped = true
		v = fmt.Sprintf(`"%s"`, v)
	}
	if v == "" {
		v = `""`
	}
	return v
}

func (doc *document) CalculateMaxFieldLengths() {
	for _, line := range doc.lines {
		if line == nil {
			continue
		}
		for fieldInd, field := range line.fields {
			fw := CalculateFieldLength(field)
			if cw, ok := line.doc.MaxColumnWidth[fieldInd]; ok {
				if cw < fw {
					line.doc.MaxColumnWidth[fieldInd] = fw
				}
			} else {
				line.doc.MaxColumnWidth[fieldInd] = fw
			}
		}
	}
}

// determine if tabular document line is valid based on the number of lines of the first row/header, returns true, nil if has the correct number of data fields
// returns false, and an error documenting the difference
func (line *documentLine) Validate() (bool, error) {
	if !line.doc.Tabular {
		return true, nil
	}
	if line.doc.HasHeaders && len(line.doc.Headers) != line.fieldCount {
		return false, fmt.Errorf("line %d does not have the correct number of fields %d/%d (current/expected)", line.line, line.fieldCount, len(line.doc.Headers))
	}
	if !line.doc.HasHeaders && line.line > 1 && len(line.doc.lines) >= 1 && len(line.doc.lines[0].fields) != line.fieldCount {
		return false, fmt.Errorf("line %d does not have the correct number of fields %d/%d (current/expected)", line.line, line.fieldCount, len(line.doc.lines[0].fields))
	}
	return true, nil
}

func (line *documentLine) Append(val string) error {
	field := RecordField{Value: val}
	if line.doc.HasHeaders && line.line == 1 {
		field.IsHeader = true
		field.FieldName = val
	}
	fieldInd := len(line.fields)
	err := checkFieldIndex(line, fieldInd)
	if err != nil {
		return ErrFieldCount
	}

	if line.line > 1 && len(line.doc.Headers)-1 > fieldInd {
		field.FieldName = line.doc.Headers[fieldInd]
	}
	field.FieldIndex = fieldInd
	line.fields = append(line.fields, field)
	fw := CalculateFieldLength(field)
	if cw, ok := line.doc.MaxColumnWidth[fieldInd]; ok {
		if cw < fw {
			line.doc.MaxColumnWidth[fieldInd] = fw
		}
	} else {
		line.doc.MaxColumnWidth[fieldInd] = fw
	}
	if line.doc.HasHeaders && line.line == 1 {
		line.doc.Headers = append(line.doc.Headers, val)
	}
	// increment the field count for the line
	line.fieldCount++
	return nil
}

func (line *documentLine) AppendNull() error {
	field := RecordField{IsNull: true}
	if line.doc.HasHeaders && line.line == 1 {
		field.IsHeader = true
		field.FieldName = "-"
	}
	fieldInd := len(line.fields)
	err := checkFieldIndex(line, fieldInd)
	if err != nil {
		return ErrFieldCount
	}

	if line.line > 1 && len(line.doc.Headers)-1 > fieldInd {
		field.FieldName = line.doc.Headers[fieldInd]
	}
	field.FieldIndex = fieldInd
	line.fields = append(line.fields, field)
	fw := CalculateFieldLength(field)
	if cw, ok := line.doc.MaxColumnWidth[fieldInd]; ok {
		if cw < fw {
			line.doc.MaxColumnWidth[fieldInd] = fw
		}
	} else {
		line.doc.MaxColumnWidth[fieldInd] = fw
	}
	if line.doc.HasHeaders && line.line == 1 {
		line.doc.Headers = append(line.doc.Headers, "-")
	}
	// increment the field count for the line
	line.fieldCount++
	return nil
}

// check field index is valid, returns the number of fields left, -1 is returned when document is not tabular or is the first line
func checkFieldIndex(line *documentLine, fieldInd int) error {
	if !line.doc.Tabular || line.line <= 1 {
		// if the document is not tabular or this is the first line this check won't be in effect
		return nil
	}
	if len(line.doc.lines) <= 1 {
		// unexpected
		return ErrNotEnoughLines
	}
	fc := line.doc.lines[0].fieldCount - fieldInd

	if fc <= 0 {
		return &WriteError{err: ErrFieldCount, line: line.line, fieldIndex: fieldInd}
	}
	return nil
}

func (line *documentLine) NextField() (*RecordField, error) {
	if len(line.fields)-1 < line.doc.currentField {
		return nil, ErrFieldNotFound
	}
	fieldInd := line.doc.currentField
	line.doc.currentField++
	return &line.fields[fieldInd], nil
}

// Returns the number of data fields, non-comment fields
func (line *documentLine) FieldCount() int {
	return line.fieldCount
}

func (line *documentLine) Line() int {
	return line.line
}

func (line *documentLine) updateHeader(fieldIndex int, val string) error {
	if !line.doc.HasHeaders {
		return nil
	}
	if line.line != 1 {
		return ErrLineIsNotHeader
	}
	if len(line.doc.Headers)-1 < fieldIndex || fieldIndex < 0 {
		return ErrFieldNotFound
	}
	line.doc.Headers[fieldIndex] = val
	for _, l := range line.doc.lines {
		for fi, field := range l.fields {
			if fi != fieldIndex {
				continue
			}
			field.FieldName = val
			l.fields[fi] = field
		}
	}
	return nil
}

// Update field at line `fi`, `fi` is 0-index
func (line *documentLine) UpdateField(fieldIndex int, val string) error {
	if len(line.fields)-1 < fieldIndex || fieldIndex < 0 {
		return ErrFieldNotFound
	}
	if line.line == 1 && line.doc.HasHeaders {
		err := line.updateHeader(fieldIndex, val)
		if err != nil {
			return err
		}
	}
	field := line.fields[fieldIndex]
	field.Value = val
	line.fields[fieldIndex] = field
	fw := CalculateFieldLength(field)
	if cw, ok := line.doc.MaxColumnWidth[fieldIndex]; ok {
		if cw < fw {
			line.doc.MaxColumnWidth[fieldIndex] = fw
		}
	} else {
		line.doc.MaxColumnWidth[fieldIndex] = fw
	}
	return nil
}
