// Create a document in the whitespace separated format.
package document

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/internetcalifornia/wsv/v2/record"
	"github.com/internetcalifornia/wsv/v2/utils"
)

type WriteError struct {
	line        int
	fieldIndex  int
	headerCount int

	err error
}

var (
	ErrFieldNotFound      = errors.New("field does not exist")
	ErrStartedToWrite     = errors.New("document started to write, need to reset document to edit")
	ErrInvalidPaddingRune = errors.New("only whitespace characters can be used for padding")
	ErrOmitHeaders        = errors.New("document configured to omit headers")
	ErrLineNotFound       = errors.New("line does not exist")
	ErrFieldCount         = errors.New("wrong number of fields")
)

func (e *WriteError) Error() string {

	if e.err == ErrStartedToWrite {
		return fmt.Sprintf("the writer already started and has currently written up until line %d, reset the writer to continue editing the document.", e.line)
	}

	if e.err == ErrFieldNotFound {
		return fmt.Sprintf("field %d, %s for line %d", e.fieldIndex, e.err.Error(), e.line)
	}

	return e.err.Error()
}

type Document interface {
	// Set the padding between values to be any from 24 unicode whitespace values, except for LineFeed. see `wsv/v1/internal` for examples
	//
	// If an invalid rune is provided an error will be returned and the Document will not change it's padding
	SetPadding(rs []rune) error
	// Add a line to the document
	//
	// Returns an error if `Write()` has been called, to continue editing after calling `Write()` first call `ResetWrite()`
	AddLine() (DocumentLine, error)
	// Returns the `DocumentLine` or an error if the line at the index does not exist.
	//
	// Lines are 1-indexed
	Line(ln int) (DocumentLine, error)
	// A convenience function for appending values to a line. Same as calling `Append(val)` and `AppendNull()` on the DocumentLine'
	//
	// Use the `Field(val)` and `Null()` functions to generate the structs for this function
	AppendLine(fields ...appendLineField) (DocumentLine, error)
	// Reset the writer to the start of the document and allow the document to be modified with `AddLine()` and DocumentLine functions
	ResetWrite()
	// Write each line as WSV encoded value in a line. Calling `Write()` advances the pointer to next line in the document.
	//
	// `Write()` is the serialized values and comments of the DocumentLine and suffixed with the line feed character
	//
	// I `Write()` is called at the end of the document the error returned is io.EOF
	Write() ([]byte, error)
	// Convenience method that calls `Write()` until the end of the document and returns the results as a slice of bytes of each line in the Document
	WriteAll() ([]byte, error)
	// Returns the number of lines in a document
	LineCount() int
	// Returns a comment for the line at the given index or an error if the lines does not have a comment
	// Line is 1-indexed
	CommentFor(ln int) (string, error)
	// Describes if the document is tabular, e.i. no succeeding lines can have fewer or more than the one it is preceded by
	//
	// Field count is determined by the 1st line (header) if document is tabular
	Tabular() bool
	// Update style of a document to be tabular or not
	SetTabularStyle(tabular bool)
	// Defines if the document has headers or not
	HasHeaders() bool
	// A method for setting the column width of a column, this is used to determine the padding needed to align the document when `Write()` is called
	SetMaxColumnWidth(col int, len int)
	// Returns the columns max width as set by `SetMaxColumnWidth()` or returns an error if there is no value in the map
	MaxColumnWidth(col int) (int, error)
	// Update the document to hide of show line-1 of a document, if document does not have a header this is a no-op
	SetHideHeaderStyle(v bool)
	// the value to show or hide the line-1 of a document if the document has headers
	EmitHeaders() bool
	// the headers of a document if the document has headers, empty if the document has no headers
	Headers() []string
	// Add headers to a document
	AppendHeader(header string)
}

type document struct {
	tabular          bool
	emitHeaders      bool
	lines            []*documentLine
	maxColumnWidth   map[int]int
	padding          []rune
	currentWriteLine int
	currentField     int
	startedWriting   bool
	headers          []string
	headerLine       int
	hasHeaders       bool
}

func (doc *document) SetPadding(rs []rune) error {
	for _, r := range rs {
		if !utils.IsFieldDelimiter(r) {
			return &WriteError{err: ErrInvalidPaddingRune}
		}
	}
	doc.padding = rs
	return nil
}

type appendLineField struct {
	val    string
	isNull bool
}

func Field(val string) appendLineField {
	return appendLineField{val, false}
}

func Null() appendLineField {
	return appendLineField{"", true}
}

func (doc *document) AppendLine(fields ...appendLineField) (DocumentLine, error) {
	line, err := doc.AddLine()
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		if field.isNull {
			err = line.AppendNull()
			if err != nil {
				return line, err
			}
			continue
		}
		err = line.Append(field.val)
		if err != nil {
			return line, err
		}
	}
	return line, nil
}

func (doc *document) AddLine() (DocumentLine, error) {
	if doc.startedWriting {
		return nil, &WriteError{err: ErrStartedToWrite, line: doc.currentWriteLine}
	}
	pln := len(doc.lines)
	line := documentLine{
		doc:    doc,
		fields: make([]record.RecordField, 0),
		line:   pln + 1,
	}

	doc.lines = append(doc.lines, &line)
	doc.currentField = 0
	return &line, nil
}

// Returns the document at the ln specified. Lines are 1-index. If the line does not exist there is an
// ErrLineNotFound error
func (doc *document) Line(ln int) (DocumentLine, error) {
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

// Write, writes the currently line to a slice of bytes based on the current line in process, calling write will increment the counter after each successful call.
// Once all lines are process will return will return empty slice, EOF
func (doc *document) Write() ([]byte, error) {
	doc.startedWriting = true
	buf := make([]byte, 0)

	if len(doc.lines)-1 < doc.currentWriteLine {
		return buf, io.EOF
	}

	line := doc.lines[doc.currentWriteLine]
	if doc.HasHeaders() && !doc.emitHeaders && doc.currentWriteLine == doc.headerLine {
		return buf, ErrOmitHeaders
	}
	// if configured to be tabular, not an empty line, and has too little/many fields compared to headers return an error
	if doc.Tabular() && doc.currentWriteLine != 0 && line.fieldCount != 0 && line.fieldCount != len(doc.Headers()) {
		return buf, &WriteError{line: line.line, headerCount: len(doc.Headers()), fieldIndex: line.fieldCount, err: ErrFieldCount}
	}

	for i, field := range line.fields {
		mw, err := doc.MaxColumnWidth(i)
		if err != nil {
			continue
		}
		v := field.SerializeText()
		p := utf8.RuneCountInString(v)
		if doc.Tabular() && (len(line.fields)-1 != i) {
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
			buf = append(buf, utils.RuneToBytes(doc.padding)...)
			buf = append(buf, []byte(v)...)
		}
	}
	if len(line.Comment()) > 0 {
		if len(buf) > 0 {
			buf = append(buf, utils.RuneToBytes(doc.padding)...)
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment()))...)
		} else {
			buf = append(buf, []byte(fmt.Sprintf("#%s", line.Comment()))...)

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

// Returns a comment if one exists for the rows or an error if comment does not exist
// lines are 1-indexed
func (doc *document) CommentFor(ln int) (string, error) {
	if len(doc.lines) < ln {
		return "", fmt.Errorf("there are no records found for row %d, please ensure you are indexing as 1-indexed values", ln)
	}
	line := doc.lines[ln-1]

	if len(line.Comment()) > 0 {
		return line.Comment(), nil
	}
	msg := fmt.Errorf("comment not found for row %d", ln)
	return "", msg
}

func (doc *document) CalculateMaxFieldLengths() {
	for _, line := range doc.lines {
		if line == nil {
			continue
		}
		for fieldInd, field := range line.fields {
			fw := field.CalculateFieldLength()
			doc.SetMaxColumnWidth(fieldInd, fw)
		}
	}
}

func (doc *document) Tabular() bool {
	return doc.tabular
}

func (doc *document) SetTabularStyle(tabular bool) {
	doc.tabular = tabular
}

func (doc *document) HasHeaders() bool {
	return doc.hasHeaders
}

func (doc *document) SetMaxColumnWidth(col int, len int) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		doc.maxColumnWidth[col] = len
		return
	}
	if v < len {
		doc.maxColumnWidth[col] = len
	}
}

func (doc *document) MaxColumnWidth(col int) (int, error) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		return 0, ErrFieldNotFound
	}
	return v, nil
}

func (doc *document) EmitHeaders() bool {
	return doc.emitHeaders
}

func (doc *document) SetHideHeaderStyle(v bool) {
	if !doc.hasHeaders {
		return
	}
	doc.emitHeaders = v
}

func (doc *document) Headers() []string {
	return doc.headers
}

func (doc *document) UpdateHeader(fi int, val string) error {
	if !doc.HasHeaders() {
		return nil
	}

	for i := range doc.LineCount() {
		line, _ := doc.Line(i)
		for fi := range line.FieldCount() {
			if fi == 0 {
				line.UpdateField(fi, val)
			}
			err := line.UpdateFieldName(fi, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (doc *document) AppendHeader(val string) {
	doc.headers = append(doc.headers, val)
}

func NewDocument() Document {
	doc := document{
		tabular:          true,
		emitHeaders:      true,
		lines:            make([]*documentLine, 0),
		currentWriteLine: 0,
		currentField:     0,
		maxColumnWidth:   make(map[int]int, 0),
		headerLine:       0,
		startedWriting:   false,
		// The runes in between data values
		padding:    []rune{' ', ' '},
		headers:    make([]string, 0),
		hasHeaders: true,
	}
	return &doc
}
