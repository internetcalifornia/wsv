package wsv

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	charCharacterTabulation     = 0x0009
	charLineFeed                = 0x000A
	charLineTabulation          = 0x000B
	charFormFeed                = 0x000C
	charCarriageReturn          = 0x000D
	charSpace                   = 0x0020
	charNextLine                = 0x0085
	charNoBreakSpace            = 0x00A0
	charOghamSpaceMark          = 0x1680
	charEnQuad                  = 0x2000
	charEmQuad                  = 0x2001
	charEnSpace                 = 0x2002
	charEmSpace                 = 0x2003
	charThreePerEmSpace         = 0x2004
	charFourPerEmSpace          = 0x2005
	charSixPerEmSpace           = 0x2006
	charFigureSpace             = 0x2007
	charPunctuationSpace        = 0x2008
	charThinSpace               = 0x2009
	charHairSpace               = 0x200A
	charLineSeparator           = 0x2028
	charParagraphSeparator      = 0x2029
	charNarrowNoBreakSpace      = 0x202F
	charMediumMathematicalSpace = 0x205F
	charIdeographicSpace        = 0x3000
)

// A ParseError is returned for parsing errors.
// Line numbers are 1-indexed and columns are 0-indexed.
type ParseError struct {
	Line          int   // Line where the error occurred
	Column        int   // Column (1-based byte index) where the error occurred
	Err           error // The actual error
	NeighborBytes []byte
}

func (e *ParseError) Error() string {
	if e.Err == ErrFieldCount {
		return fmt.Sprintf("record on line %d: %v", e.Line, e.Err)
	}
	return fmt.Sprintf("parse error on line %d, column %d [%s]: %v", e.Line, e.Column, string(e.NeighborBytes), e.Err)

}

// These are the errors that can be returned in ParseError.Err.
var (
	ErrBareQuote    = errors.New("bare \" in non-quoted-field")
	ErrQuote        = errors.New("extraneous or missing \" in quoted-field")
	ErrFieldCount   = errors.New("wrong number of fields")
	ErrLineFeedTerm = errors.New("line feed terminated before the line end end")
	ErrInvalidNull  = errors.New("null `-` specifier cannot be included without white space surrounding, unless it is the last value in the line. To record a literal `-` please wrap the value in double quotes")
)

type Reader struct {
	numLine             int
	offset              int64
	rawBuffer           []byte
	FieldsPerRecord     int
	Comments            map[int]*RecordField
	headers             []string
	IncludesHeader      bool
	IsTabular           bool
	r                   *bufio.Reader
	NullTrailingColumns bool
}

func (r *Reader) Headers() []string {
	return r.headers
}

type RecordField struct {
	reader     *Reader
	IsNull     bool
	Value      string
	FieldIndex int
	RowIndex   int
	FieldName  string
	IsHeader   bool
	IsComment  bool
}

// Returns a comment if one exists for the rows or returns an error
// Rows are 1-indexed
func (r *Reader) CommentFor(row int) (string, error) {

	for r, field := range r.Comments {
		if r == row && field.IsComment {
			return field.Value, nil
		}
	}
	msg := fmt.Errorf("comment not found for row %d", row)
	return "", msg
}

func getIndexOfSlice[T any](s []T, i int) (*T, error) {
	if i < 0 {
		return nil, errors.New("index must be be 0 or greater")
	}
	if i > len(s)-1 {
		message := fmt.Sprintf("index %d is greater than %d", i, len(s)-1)
		return nil, errors.New(message)
	}
	return &s[i], nil
}

func columnName(headers []string, index int) string {
	v, err := getIndexOfSlice(headers, index)
	if err != nil {
		return ""
	}
	return strings.Clone(*v)
}

func isFieldDelimiter(rn rune) bool {
	return (rn == charCharacterTabulation ||
		rn == charLineTabulation ||
		rn == charFormFeed ||
		rn == charCarriageReturn ||
		rn == charSpace ||
		rn == charNextLine ||
		rn == charNoBreakSpace ||
		rn == charOghamSpaceMark ||
		rn == charEnQuad ||
		rn == charEmQuad ||
		rn == charEnSpace ||
		rn == charEmSpace ||
		rn == charThreePerEmSpace ||
		rn == charFourPerEmSpace ||
		rn == charSixPerEmSpace ||
		rn == charFigureSpace ||
		rn == charPunctuationSpace ||
		rn == charThinSpace ||
		rn == charHairSpace ||
		rn == charLineSeparator ||
		rn == charParagraphSeparator ||
		rn == charNarrowNoBreakSpace ||
		rn == charMediumMathematicalSpace ||
		rn == charIdeographicSpace)
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:                   bufio.NewReader(r),
		IsTabular:           true,
		IncludesHeader:      true,
		NullTrailingColumns: true,
		Comments:            make(map[int]*RecordField),
	}
}

// Return the column name at the index i, will return "" if not found
func (r *Reader) ColumnNameOf(i int) (*string, error) {
	return getIndexOfSlice(r.headers, i)
}

// Return the index of a column name
func (r *Reader) IndexedAt(n string) []int {
	idxs := make([]int, 0)
	for i, h := range r.headers {
		if h != n {
			continue
		}
		idxs = append(idxs, i)
	}
	return idxs

}

func Parse(wsvFile string) ([][]RecordField, error) {
	file, err := os.Open(wsvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := NewReader(file)
	records, err := r.ReadAll()
	return records, err
}

func (r *Reader) ReadAll() (records [][]RecordField, err error) {
	for {
		record, err := r.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
}

type LineField struct {
	Value     string
	IsComment bool
	IsNull    bool
}

func ParseLine(n int, line []byte) ([]LineField, error) {
	var b1 *byte = nil
	var b2 *byte = nil
	var b3 *byte = nil
	var b4 *byte = nil
	doubleQuoted := false
	endedDoubleQuote := false
	escapeNewLinePos := 0
	isNull := false
	var startDoubleQuote int = 0
	s := ""
	str := make([]LineField, 0)
lineLoop:
	for i, b0 := range line {
		if b4 != nil {
			b4 = b3
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b4 == nil && b3 != nil {
			b4 = b3
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b3 == nil && b2 != nil {
			b3 = b2
			b2 = b1
			b1 = &b0
		}
		if b2 == nil && b1 != nil {
			b2 = b1
			b1 = &b0
		}
		if b1 == nil {
			b1 = &b0
		}
		r := rune(b0)

		switch r {
		case '\n':
			if i < len(line)-1 {
				d := neighborBytes(i, line)
				return str, &ParseError{Line: n, Column: i, Err: ErrLineFeedTerm, NeighborBytes: d}
			}
			break lineLoop
		case '#':
			if !doubleQuoted {
				s = s + string(line[i:])
				// since we are copying to the end of line we should remove the suffix of the line feed
				s = strings.TrimSuffix(s, string('\n'))
				str = append(str, LineField{IsComment: true, Value: s, IsNull: isNull})
				s = ""
				break lineLoop
			}
			s = s + string(r)
			continue
		case '"':

			if bytesToString(b3, b2, b1) == `"/"` {
				s = strings.TrimSuffix(s, string('/')) + string('\n')
				escapeNewLinePos = i
				// escape new line
				continue
			}
			if i > escapeNewLinePos && b2 != nil && rune(*b2) == '"' && (b3 == nil || rune(*b3) != '"') {
				if len(line)-1 > i+1 && isFieldDelimiter(nextRune(line[i+1:])) && b3 != nil && rune(*b3) == '/' {
					// edge case field ends with newline
				} else if len(line)-1 > i+2 && nextRune(line[i+1:]) == '/' && nextRune(line[i+2:]) == '"' {
					// edge case field starts with newline
				} else {
					s = s + `"`
					continue
				}
			}
			if doubleQuoted && ((len(line)-1 == i) || (len(line)-1 > i && isFieldDelimiter(nextRune(line[i+1:])))) {
				doubleQuoted = false
				endedDoubleQuote = true
			}
		case '-':
			if r == '-' && (b2 == nil || isFieldDelimiter(rune(*b2))) && !doubleQuoted {
				isNull = true
			}
			fallthrough
		default:
			if bytesToString(b3, b2, b1) == `"/"` {
				s = strings.TrimSuffix(s, string('/')) + string('\n')
			}
			if isNull && (len(line)-1 == i) {
				str = append(str, LineField{IsComment: false, Value: "", IsNull: isNull})
				break lineLoop
			}
			// currently flagged as null but has more characters left to parse and the next immediate character is not a white space or the end of the string, and is not surround by double quotes we have an invalid
			if isNull && len(line)-1 > i && bytes.IndexFunc(line[i:], isFieldDelimiter) != 1 {
				if b2 != nil && rune(*b2) == '-' && bytes.IndexFunc([]byte{*b1}, isFieldDelimiter) == 0 {
					s = ""
				} else {
					return str, &ParseError{Column: i, Err: ErrInvalidNull}
				}
			}
			if (b3 == nil || rune(*b3) != '"') && (b2 != nil && rune(*b2) == '"') && startDoubleQuote <= i {
				if strings.HasSuffix(bytesToString(b4, b3, b2, b1), `"/`) ||
					strings.HasPrefix(bytesToString(b4, b3, b2, b1), `"/"`) {
				} else {
					startDoubleQuote = i - 1
					doubleQuoted = true
				}
			}
			isDelim := isFieldDelimiter(r)
			if isDelim && (!doubleQuoted || endedDoubleQuote) {
				doubleQuoted = false
				endedDoubleQuote = false
				str = append(str, LineField{IsComment: false, Value: s, IsNull: isNull})
				isNull = false
				s = ""
				continue
			}
			s = s + string(r)
			continue
		}
	}
	if doubleQuoted {
		// the following string value could not be parsed correctly
		nb := neighborBytes(startDoubleQuote, line)
		return str, &ParseError{Column: startDoubleQuote, Err: ErrBareQuote, Line: n, NeighborBytes: nb}
	}
	if len(s) > 0 {
		str = append(str, LineField{IsComment: false, Value: s, IsNull: isNull})
	}
	return str, nil
}

func neighborBytes(i int, line []byte) (neighbor []byte) {
	if i < 0 {
		return neighbor
	}
	p := 5 - (5 - i)
	s := 5 - (5 - i)
	if p > 5 {
		p = 5
	}
	if s > 5 {
		s = 5
	}
	if len(line[i:]) < i+s {
		s = (len(line[i:]) - 1) + i
	}
	neighbor = line[p:s]
	return neighbor
}

func bytesToString(s ...*byte) string {
	str := ""
	for _, b := range s {
		if b == nil {
			continue
		}
		str = str + string(*b)
	}
	return str
}

func (r *Reader) CurrentRow() int {
	return r.numLine
}

// Read a slice of RecordField from r.
// If the reader IsTabular and the row being parsed has more
// fields than the header row will return the records, ParseError
// Read returns the record along with the error ParseError.
// If the record contains a field that cannot be parsed,
// Read returns a partial record along with the parse error.
// The partial record contains all fields read before the error.
// If there is no data left to be read, Read returns an empty RecordField slice, io.EOF.
func (r *Reader) Read() ([]RecordField, error) {
	var line []byte
	var errRead error
	records := make([]RecordField, 0)
	line, errRead = r.readLine()
	if errRead == io.EOF {
		return records, io.EOF
	}
	fields, errRead := ParseLine(r.numLine, line)
	if errRead != nil {
		return records, errRead
	}
	for i, field := range fields {
		if r.numLine == 1 && r.IncludesHeader && !field.IsComment {
			r.headers = append(r.headers, field.Value)
			d := RecordField{reader: r, IsNull: false, Value: field.Value, RowIndex: r.numLine, FieldIndex: i, IsHeader: true, FieldName: field.Value, IsComment: false}
			records = append(records, d)
			continue
		}
		if field.IsComment {
			d := RecordField{reader: r, IsNull: false, Value: field.Value, RowIndex: r.numLine, FieldIndex: i, IsHeader: false, FieldName: field.Value, IsComment: true}
			r.Comments[r.numLine] = &d
			records = append(records, d)
			continue
		}
		recCount := lenOfRecordsWithoutComments(records)
		if r.IsTabular && r.IncludesHeader && len(r.headers) < recCount {
			return records, &ParseError{Line: r.numLine, Column: 0, Err: ErrFieldCount}
		}
		fieldName := columnName(r.headers, i)
		d := RecordField{reader: r, IsNull: false, Value: field.Value, RowIndex: r.numLine, FieldIndex: i, IsHeader: false, FieldName: fieldName}
		records = append(records, d)
	}

	if len(records) == 0 {
		return records, errRead
	}

	if len(records) == 1 && records[0].IsComment {
		return records, errRead
	}

	if r.numLine != 1 && r.NullTrailingColumns && len(records) < len(r.headers) {
		x := len(r.headers) - len(records)
		o := len(records)
		for i := range x {
			h := o + i
			cname := columnName(r.headers, h)
			rec := RecordField{IsNull: true, Value: "", FieldIndex: h, RowIndex: r.numLine, FieldName: cname, IsHeader: false}
			records = append(records, rec)
		}
	}

	return records, errRead

}

func lenOfRecordsWithoutComments(r []RecordField) int {
	c := 0
	for _, rec := range r {
		if rec.IsComment {
			continue
		}
		c++
	}
	return c
}

func nextRune(b []byte) rune {
	r, _ := utf8.DecodeRune(b)
	return r
}

func (r *Reader) readLine() ([]byte, error) {
	line, err := r.r.ReadSlice(charLineFeed)
	if err == bufio.ErrBufferFull {
		r.rawBuffer = append(r.rawBuffer[:0], line...)
		for err == bufio.ErrBufferFull {
			line, err = r.r.ReadSlice(charLineFeed)
			r.rawBuffer = append(r.rawBuffer, line...)
		}
		line = r.rawBuffer
	}
	readSize := len(line)
	if readSize > 0 && err == io.EOF {
		err = nil
		// For backwards compatibility, drop trailing \r before EOF.
		if line[readSize-1] == charCarriageReturn {
			line = line[:readSize-1]
		}
	}
	r.numLine++
	r.offset += int64(readSize)
	if n := len(line); n >= 2 && line[n-2] == charCarriageReturn && line[n-1] == charLineFeed {
		line[n-2] = charLineFeed
		line = line[:n-1]
	}
	// trim the trailing new line
	line = bytes.TrimSuffix(line, []byte("\n"))
	return line, err
}
