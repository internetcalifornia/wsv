// Create a document in the whitespace separated format.
package document

import (
	"errors"
	"fmt"
	"io"
	"strings"
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

type Document struct {
	Tabular          bool
	EmitHeaders      bool
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

func (doc *Document) SetPadding(rs []rune) error {
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

func (doc *Document) AppendLine(fields ...appendLineField) (DocumentLine, error) {
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

func (doc *Document) AddLine() (DocumentLine, error) {
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
func (doc *Document) Line(ln int) (DocumentLine, error) {
	if len(doc.lines)-1 < ln-1 || ln < 1 {
		return nil, ErrLineNotFound
	}
	line := doc.lines[ln-1]
	return line, nil
}

func (doc *Document) ResetWrite() {
	doc.startedWriting = false
	doc.currentWriteLine = 0
}

func (doc *Document) WriteLine(n int, includeHeader bool) ([]byte, error) {
	buf := make([]byte, 0)
	if n > len(doc.lines) || n < 1 {
		return buf, ErrLineNotFound
	}

	headers, err := doc.Line(doc.headerLine)
	if err != nil && includeHeader {
		return buf, err
	}
	fmt.Printf("Headers %d %+v\n", doc.headerLine, headers)
	line := doc.lines[n-1]
	fmt.Printf("Line %+v\n", line)
	headerLine := make([]string, line.FieldCount())
	dataLine := make([]string, line.FieldCount())
	for i, field := range line.fields {
		var header = ""
		if headers != nil {
			headerField, err := headers.Field(i)
			if err == nil {
				header = headerField.SerializeText()
			}
		}
		data := field.SerializeText()
		dl := utf8.RuneCountInString(data)
		hl := utf8.RuneCountInString(header)
		if includeHeader && i < line.fieldCount-1 {
			if dl >= hl {
				for range dl - hl {
					header = header + " "
				}
			} else {
				for range hl - dl {
					data = data + " "
				}
			}
		}
		headerLine[i] = header
		dataLine[i] = data
	}

	if includeHeader {
		return []byte(strings.Join(headerLine, string(doc.padding)) + "\n" + strings.Join(dataLine, string(doc.padding))), nil
	}
	return []byte(strings.Join(dataLine, string(doc.padding))), nil
}

// Write, writes the currently line to a slice of bytes based on the current line in process, calling write will increment the counter after each successful call.
// Once all lines are process will return will return empty slice, EOF
func (doc *Document) Write() ([]byte, error) {
	doc.startedWriting = true
	buf := make([]byte, 0)

	if len(doc.lines)-1 < doc.currentWriteLine {
		return buf, io.EOF
	}

	line := doc.lines[doc.currentWriteLine]
	if doc.HasHeaders() && !doc.EmitHeaders && doc.currentWriteLine == doc.headerLine {
		return buf, ErrOmitHeaders
	}
	// if configured to be tabular, not an empty line, and has too little/many fields compared to headers return an error
	if doc.Tabular && doc.currentWriteLine != 0 && line.fieldCount != 0 && line.fieldCount != len(doc.Headers()) {
		return buf, &WriteError{line: line.line, headerCount: len(doc.Headers()), fieldIndex: line.fieldCount, err: ErrFieldCount}
	}

	for i, field := range line.fields {
		mw, err := doc.MaxColumnWidth(i)
		if err != nil {
			continue
		}
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

func (doc *Document) WriteAll() ([]byte, error) {
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

func (doc *Document) LineCount() int {
	return len(doc.lines)
}

// Returns a comment if one exists for the rows or an error if comment does not exist
// lines are 1-indexed
func (doc *Document) CommentFor(ln int) (string, error) {
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

func (doc *Document) CalculateMaxFieldLengths() {
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

func (doc *Document) HasHeaders() bool {
	return doc.hasHeaders
}

func (doc *Document) SetMaxColumnWidth(col int, len int) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		doc.maxColumnWidth[col] = len
		return
	}
	if v < len {
		doc.maxColumnWidth[col] = len
	}
}

func (doc *Document) MaxColumnWidth(col int) (int, error) {
	v, ok := doc.maxColumnWidth[col]
	if !ok {
		return 0, ErrFieldNotFound
	}
	return v, nil
}

func (doc *Document) SetHideHeaderStyle(v bool) {
	if !doc.hasHeaders {
		return
	}
	doc.EmitHeaders = v
}

func (doc *Document) Headers() []string {
	return doc.headers
}

func (doc *Document) UpdateHeader(fi int, val string) error {
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

func (doc *Document) AppendHeader(val string) {
	doc.headers = append(doc.headers, val)
}

func NewDocument() *Document {
	doc := Document{
		Tabular:          true,
		EmitHeaders:      true,
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
