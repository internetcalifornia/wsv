package wsv

import (
	"errors"
	"fmt"
	"strings"
)

type Writer struct {
	Tabular        bool
	EmitHeaders    bool
	Records        [][]RecordField
	maxColumnWidth map[int]int
	currentIndex   int
	Headers        []string
	HasHeaders     bool
}

func NewWriter() *Writer {
	w := Writer{
		Tabular:        true,
		EmitHeaders:    true,
		Records:        make([][]RecordField, 0),
		currentIndex:   0,
		maxColumnWidth: make(map[int]int, 0),
		Headers:        make([]string, 0),
		HasHeaders:     true,
	}
	return &w
}

func (w *Writer) Append(rowIdx int, f RecordField) error {
	if rowIdx < 0 {
		return errors.New("cannot append a record field to a non-positive integer")
	}
	rowCount := len(w.Records)
	// attempting to insert field data into row
	if rowCount < rowIdx {
		// currently there are only
		for {
			if rowCount >= rowIdx {
				// we have enough rows now
				break
			}
			// append blank row for row
			w.Records = append(w.Records, []RecordField{})
			rowCount++
		}
	}
	// establish a pointer to the slice of record fields
	var rowData *[]RecordField = nil
	// search for the row
	for rowNum := range w.Records {
		if rowNum != rowIdx {
			continue
		}
		rd := w.Records[rowNum]
		rowData = &rd
		// get the address to to
		break
	}
	if rowData == nil {
		// the slice of records does not exists for row
		v := calculateFieldLength(f)
		if cw, ok := w.maxColumnWidth[f.FieldIndex]; ok {
			if cw < len(v) {
				w.maxColumnWidth[rowIdx] = len(v)
			}
		} else {
			w.maxColumnWidth[f.FieldIndex] = len(v)
		}
		if rowIdx == 0 && w.HasHeaders && f.IsHeader {
			w.Headers = append(w.Headers, f.Value)
		}
		w.Records = append(w.Records, []RecordField{f})
	} else {
		v := calculateFieldLength(f)

		if cw, ok := w.maxColumnWidth[f.FieldIndex]; ok {
			if cw < len(v) {
				w.maxColumnWidth[rowIdx] = len(v)
			}
		} else {
			w.maxColumnWidth[f.FieldIndex] = len(v)
		}
		d := *rowData
		d = append(d, f)
		if rowIdx == 0 && w.HasHeaders && f.IsHeader {
			w.Headers = append(w.Headers, f.Value)
		}
		w.Records[rowIdx] = d
	}
	return nil
}

func calculateFieldLength(f RecordField) string {
	v := f.Value
	v = strings.ReplaceAll(v, `"`, `""`)
	v = strings.ReplaceAll(v, "\n", `"/"`)
	if strings.ContainsFunc(v, isFieldDelimiter) {
		v = fmt.Sprintf(`"%s"`, v)
	}
	return v
}

type WriteOptions struct {
	MaxColumnWidth map[int]int
	HasHeaders     bool
	Tabular        bool
	Rows           [][]RecordField
	EmitHeaders    bool
}

func ToString(opt WriteOptions) string {
	str := ""

	for x, r := range opt.Rows {
		if opt.HasHeaders && !opt.EmitHeaders && x == 0 {
			continue
		}
		for i, f := range r {
			mw := opt.MaxColumnWidth[i]
			v := f.SerializeText()
			p := len(f.Value)
			if opt.Tabular {
				for {
					// pad value with single spaces unless it's the last column
					if p < mw && len(r)-1 != i {
						v = fmt.Sprintf("%s%s", v, " ")
						p = len(v)
						continue
					}
					break
				}
			}
			if i == 0 {
				str = fmt.Sprintf("%s%s", str, v)
			} else {
				str = fmt.Sprintf("%s %s", str, v)
			}
		}
		str = fmt.Sprintf("%s\n", str)
	}
	return str
}

func (w *Writer) Write() string {
	opt := WriteOptions{EmitHeaders: w.EmitHeaders, HasHeaders: w.HasHeaders, Rows: w.Records, Tabular: w.Tabular, MaxColumnWidth: w.maxColumnWidth}
	return ToString(opt)
}

// Returns a comment if one exists for the rows or an error if comment does not exist
// Rows are 1-indexed
func (w *Writer) CommentFor(row int) (string, error) {
	if len(w.Records) < row {
		return "", errors.New("there are no records found for row %d, please ensure you are indexing as 1-indexed values")
	}
	fields := w.Records[row-1]

	for _, field := range fields {
		if field.IsComment {
			return field.Value, nil
		}
	}
	msg := fmt.Errorf("comment not found for row %d", row)
	return "", msg
}

func (f *RecordField) SerializeText() string {
	if f.IsNull {
		return "-"
	}
	if f.IsComment {
		return f.Value
	}
	v := f.Value
	v = strings.ReplaceAll(v, `"`, `""`)
	if strings.Contains(v, "\n") {
		v = fmt.Sprintf(`"%s"`, v)
	}
	v = strings.ReplaceAll(v, "\n", `"/"`)
	if strings.ContainsFunc(v, isFieldDelimiter) {
		v = fmt.Sprintf(`"%s"`, v)
	}
	return v
}
