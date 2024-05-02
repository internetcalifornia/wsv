package document

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/internetcalifornia/wsv/v1/internal"
)

type WriteError struct {
	line        int
	fieldIndex  int
	headerCount int

	err error
}

func (e *WriteError) Error() string {

	if e.err == internal.ErrFieldCount {
		return fmt.Sprintf("trying to write a value on line %d for the field index %d is not allowed because the document has %d header(s) and the writer is configured to be tabular.", e.line, e.fieldIndex, e.headerCount)
	}

	if e.err == internal.ErrStartedToWrite {
		return fmt.Sprintf("the writer already started and has currently written up until line %d, reset the writer to continue editing the document.", e.line)
	}

	if e.err == internal.ErrLineNotEditable {
		return fmt.Sprintf("line %d is non-editable %s", e.line, e.err.Error())
	}

	if e.err == internal.ErrFieldNotFound {
		return fmt.Sprintf("field %d, %s for line %d", e.fieldIndex, e.err.Error(), e.line)
	}

	return e.err.Error()
}

type Document interface {
	SetPadding(rs []rune) error
	AddLine() (DocumentLine, error)
	Line(ln int) (DocumentLine, error)
	ResetWrite()
	Write() ([]byte, error)
	WriteAll() ([]byte, error)
	LineCount() int
	CommentFor(ln int) (string, error)
	Tabular() bool
	SetTabularStyle(tabular bool)
	HasHeaders() bool
	SetMaxColumnWidth(col int, len int)
	MaxColumnWidth(col int) (int, error)
	SetHideHeaderStyle(v bool)
	EmitHeaders() bool
	Headers() []string
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
	hasHeaders       bool
}

func (doc *document) SetPadding(rs []rune) error {
	for _, r := range rs {
		if !internal.IsFieldDelimiter(r) {
			return &WriteError{err: internal.ErrInvalidPaddingRune}
		}
	}
	doc.padding = rs
	return nil
}

func (doc *document) AddLine() (DocumentLine, error) {
	if doc.startedWriting {
		return nil, &WriteError{err: internal.ErrStartedToWrite, line: doc.currentWriteLine}
	}
	pln := len(doc.lines)
	line := documentLine{
		doc:    doc,
		fields: make([]internal.RecordField, 0),
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
		return nil, internal.ErrLineNotFound
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
	if doc.HasHeaders() && !doc.emitHeaders && doc.currentWriteLine == 0 {
		return buf, internal.ErrOmitHeaders
	}
	// if configured to be tabular, not an empty line, and has too little/many fields compared to headers return an error
	if doc.Tabular() && doc.currentWriteLine != 0 && line.fieldCount != 0 && line.fieldCount != len(doc.Headers()) {
		return buf, &WriteError{line: line.line, headerCount: len(doc.Headers()), fieldIndex: line.fieldCount, err: internal.ErrFieldCount}
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
			buf = append(buf, internal.RuneToBytes(doc.padding)...)
			buf = append(buf, []byte(v)...)
		}
	}
	if len(line.Comment()) > 0 {
		if len(buf) > 0 {
			buf = append(buf, internal.RuneToBytes(doc.padding)...)
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
		return 0, internal.ErrFieldNotFound
	}
	return v, nil
}

func (doc *document) EmitHeaders() bool {
	return doc.emitHeaders
}

func (doc *document) SetHideHeaderStyle(v bool) {
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

func NewDocument() document {
	doc := document{
		tabular:          true,
		emitHeaders:      true,
		lines:            make([]*documentLine, 0),
		currentWriteLine: 0,
		currentField:     0,
		maxColumnWidth:   make(map[int]int, 0),
		startedWriting:   false,
		// The runes in between data values
		padding:    []rune{' ', ' '},
		headers:    make([]string, 0),
		hasHeaders: true,
	}
	return doc
}
