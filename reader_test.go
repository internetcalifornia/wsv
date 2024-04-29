package wsv_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/internetcalifornia/wsv"
)

func TestQuotedValuesLine(t *testing.T) {
	line := `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "Given Name" {
		t.Errorf(`expected field 1 to be Given Name but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "Family Name" {
		t.Errorf(`expected field 2 to be Family Name but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "Date of Birth" {
		t.Errorf(`expected field 3 to be Date of Birth but got %+v\n`, r[2])
	}
	if r[3].IsNull || r[3].Value != "Favorite Color" {
		t.Errorf(`expected field 3 to be Favorite Color but got %+v\n`, r[2])
	}
}

func toPointer[t any](a t) *t {
	return &a
}

func TestIsEmptyStringLiteral(t *testing.T) {
	t1 := []byte{'\t', '"', '"', '\t'}
	t1_p := []*byte{}
	for _, i := range t1 {
		t1_p = append(t1_p, toPointer(i))
	}

	t2 := []byte{'"', '"', '\t'}
	t2_p := []*byte{nil}
	for _, i := range t2 {
		t2_p = append(t2_p, toPointer(i))
	}

	t3 := []byte{'"', '"'}
	t3_p := []*byte{nil}
	for _, i := range t3 {
		t3_p = append(t3_p, toPointer(i))
	}
	t3_p = append(t3_p, nil)

	t4 := []byte{' ', '"', '"', ' '}
	t4_p := []*byte{}
	for _, i := range t4 {
		t4_p = append(t4_p, toPointer(i))
	}

	t5 := []byte{'\t', '"', '"'}
	t5_p := []*byte{}
	for _, i := range t5 {
		t5_p = append(t5_p, toPointer(i))
	}
	t5_p = append(t5_p, nil)

	if !wsv.IsLiteralEmptyString(t1_p) {
		t.Error("failed test 1")
	}
	if !wsv.IsLiteralEmptyString(t2_p) {
		t.Error("failed test 2")
	}
	if !wsv.IsLiteralEmptyString(t3_p) {
		t.Error("failed test 3")
	}
	if !wsv.IsLiteralEmptyString(t4_p) {
		t.Error("failed test 4")
	}
	if !wsv.IsLiteralEmptyString(t5_p) {
		t.Error("failed test 5")
	}
}

func TestEmptyStringLiteral(t *testing.T) {
	line := `India						""					🇮🇳			  -`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "India" {
		t.Errorf(`expected field 1 to be India but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "" {
		t.Errorf(`expected field 2 to be [EMPTY STRING] but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "🇮🇳" {
		t.Errorf(`expected field 3 to be 🇮🇳 but got %+v\n`, r[2])
	}
	if !r[3].IsNull || r[3].Value != "" {
		t.Errorf(`expected field 3 to be [NULL] but got %+v\n`, r[2])
	}
}

func TestReadColumnAndDataWithSpaces(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`)
	lines = append(lines, `"Jean Smith" "Le Croix" "Jan 01 2023" "Space Purple"`)
	lines = append(lines, `"Mary Jane" "Vasquez Rojas" "Feb 02 2021" "Midnight Grey"`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()

	if err != nil {
		t.Error(err)
	}
	if len(line.Fields) < 4 {
		t.Errorf("expected 4 fields but only got %d instead", len(line.Fields))
	}
	if !line.Fields[0].IsHeader || line.Fields[0].Value != "Given Name" {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != "Family Name" {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != "Date of Birth" {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != "Favorite Color" {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "Given Name" || line.Fields[0].Value != "Jean Smith" {
		t.Errorf("Row 1 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "Family Name" || line.Fields[1].Value != "Le Croix" {
		t.Errorf("Row 1 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "Date of Birth" || line.Fields[2].Value != "Jan 01 2023" {
		t.Errorf("Row 1 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "Favorite Color" || line.Fields[3].Value != "Space Purple" {
		t.Errorf("Row 1 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "Given Name" || line.Fields[0].Value != "Mary Jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "Family Name" || line.Fields[1].Value != "Vasquez Rojas" {
		t.Errorf("Row 2 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "Date of Birth" || line.Fields[2].Value != "Feb 02 2021" {
		t.Errorf("Row 2 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "Favorite Color" || line.Fields[3].Value != "Midnight Grey" {
		t.Errorf("Row 2 Column 4 does not match %+v", line.Fields[3])
	}
}

func TestParseDataWithEmoji(t *testing.T) {

	fields, err := wsv.ParseLine(1, []byte("France						Paris	            🇫🇷			  \"The Eiffel Tower was built for the 1889 World's Fair.\"/\"It was almost torn down afterwards.\""))
	if err != nil {
		t.Error(err)
		return

	}
	if len(fields) != 4 {
		t.Errorf("expected 4 fields but only got %d instead", len(fields))
		return
	}

	if fields[0].IsNull || fields[0].Value != "France" || fields[0].IsComment {
		t.Error("expected field 1 to be [France] but got", fields[0].Value, "instead")
	}

	if fields[1].IsNull || fields[1].Value != "Paris" || fields[1].IsComment {
		t.Error("expected field 2 to be [Paris] buParist got", fields[1].Value, "instead")
	}

	if fields[2].IsNull || fields[2].Value != "🇫🇷" || fields[2].IsComment {
		t.Error("expected field 3 to be [🇫🇷] but got", fields[2].Value, "instead")
	}

	if fields[3].IsNull || fields[3].Value != "The Eiffel Tower was built for the 1889 World's Fair."+string('\n')+"It was almost torn down afterwards." || fields[3].IsComment {
		t.Error("expected field 4 to be [The Eiffel Tower was built for the 1889 World's Fair.\\nIt was almost torn down afterwards.] but got", fields[3].Value, "instead")
	}
}

func TestReadColumnAndDataWithoutSpaces(t *testing.T) {
	lines := make([]string, 0, 3)
	lines = append(lines, `fName lName dob gender`)
	lines = append(lines, `john doe 2023-01-01 M`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.Fields[0].IsHeader || line.Fields[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line.Fields[3])
	}
}

func TestReadColumnAndDataWithComments(t *testing.T) {
	lines := make([]string, 0, 4)
	lines = append(lines, `fName lName dob gender #these are headers`)
	lines = append(lines, `john doe 2023-01-01 M #this data is probably not accurate`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	lines = append(lines, `#this is a comment`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.Fields[0].IsHeader || line.Fields[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.Comment != "this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %s", err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if len(line.Fields) != 0 && line.Comment == "this is a comment" {
		t.Error("expected the line to have a length of 0 but got", len(line.Fields), "instead")
		return
	}

	if len(line.Comment) == 0 {
		t.Error("expected this line to contain a comment but got", line.Comment, "instead")
	}
}

func TestParseWithEmptyLinesAndComments(t *testing.T) {
	lines := make([]string, 0, 4)
	lines = append(lines, `fName lName dob gender #these are headers`)
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "#here we go!")
	lines = append(lines, `john doe 2023-01-01 M #this data is probably not accurate`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	lines = append(lines, `#this is a comment`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.Fields[0].IsHeader || line.Fields[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line.Fields) != 0 {
		t.Error("expected row 2 to have zero fields but got", len(line.Fields), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line.Fields) != 0 {
		t.Error("expected row 3 to have zero fields but got", len(line.Fields), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line.Fields) > 0 {
		t.Error("expected row 4 to have 0 data field and 1 comment but got", len(line.Fields), "instead")
		return
	}
	if len(line.Comment) == 0 {
		t.Error("expected a comment but got", line.Comment, "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if line.Comment != "this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %+v", err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "john" {
		t.Errorf("Row 5 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "doe" {
		t.Errorf("Row 5 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2023-01-01" {
		t.Errorf("Row 5 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "M" {
		t.Errorf("Row 5 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != "fName" || line.Fields[0].Value != "jane" {
		t.Errorf("Row 6 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != "lName" || line.Fields[1].Value != "smith" {
		t.Errorf("Row 6 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != "dob" || line.Fields[2].Value != "2024-01-01" {
		t.Errorf("Row 6 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != "gender" || line.Fields[3].Value != "F" {
		t.Errorf("Row 6 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if len(line.Fields) != 0 {
		t.Error("expected row 7 to have a length of 1 but got", len(line.Fields), "instead")
		return
	}

	if len(line.Comment) == 0 {
		t.Error("expected row 7 to contain a comment but got", line.Comment, "instead")
	}
	if r.CurrentRow() != 7 {
		t.Error("expected current row to be 7 but got", r.CurrentRow(), "instead")
	}
}

func TestParseLineWithComments(t *testing.T) {
	line := `john "hungry ""doe""" #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLine(t *testing.T) {
	line := `john "hungry"/"doe" #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != "hungry\ndoe" {
		t.Errorf("expected field 2 to be hungry|doe but got %+v instead", r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithTrailingNewLineFollowedBySpace(t *testing.T) {
	line := `john "hungry"/"doe"/" " #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry`+string('\n')+`doe`+string('\n')+` ` {
		t.Errorf(`expected field 2 to be [hungry"\\n"doe\\n ] but got [%s] instead`, r[1].Value)
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAndEscapedDoubleQuotes(t *testing.T) {
	line := `john "hungry"""/"""doe" #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry"`+string('\n')+`"doe` {
		t.Errorf(`expected field 2 to be hungry"\\n"doe but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestQuotedValuesStartWith(t *testing.T) {
	line := `"john" "hungry" "hippo"`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john" {
		t.Errorf(`expected field 1 to be john but got %+v\n`, r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if r[2].IsNull || r[2].Value != "hippo" {
		t.Errorf(`expected field 3 to be hippo but got %+v\n`, r[2])
	}
}

func TestParseLineNullEndValue(t *testing.T) {
	line := `john hungry	-`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	// t.Errorf("%+v", r)
	if r[0].IsNull || r[0].Value != "john" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if !r[2].IsNull || r[2].Value != "" {
		t.Errorf(`expected field 3 to be null but got %+v\n`, r[2])
	}
}

func TestLiterallyDashValue(t *testing.T) {
	line := `"john-smith" hungry hippo`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john-smith" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if r[1].IsNull || r[1].Value != "hungry" {
		t.Errorf(`expected field 2 to be hungry but got %+v\n`, r[1])
	}
	if r[1].IsNull || r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseInvalidDashValue(t *testing.T) {
	line := `john -hungry	"hippo"`
	_, err := wsv.ParseLine(1, []byte(line))
	if err.Error() != "parse error on line 0, column 5 []: null `-` specifier cannot be included without white space surrounding, unless it is the last value in the line. To record a literal `-` please wrap the value in double quotes" {
		t.Error(err)
	}
}

func TestParseInvalidDashInLastField(t *testing.T) {
	line := `john hungry -hippo`
	_, err := wsv.ParseLine(1, []byte(line))
	if err == nil {
		t.Error("expected this to throw an error")
	}
}

func TestParseBareDoubleQuote(t *testing.T) {
	line := `India						"""					🇮🇳			  -`
	r, err := wsv.ParseLine(1, []byte(line))
	if err == nil {
		t.Error("Should have returned a bare double quote error", r)
		return
	}
}

func TestParseBareDoubleQuoteLastChar(t *testing.T) {
	line := `India						"""`
	r, err := wsv.ParseLine(1, []byte(line))
	if err == nil {
		t.Error("Should have returned a bare double quote error", "["+r[1].Value+"]")
		return
	}
}

func TestParseLineNullMiddleValue(t *testing.T) {
	line := `john -	hippo`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	// t.Errorf("%+v", r)
	if r[0].IsNull || r[0].Value != "john" {
		t.Error("expected field 1 to be john but got", r[0])
	}
	if !r[1].IsNull || r[1].Value != "" {
		t.Errorf(`expected field 2 to be null but got %+v\n`, r[1])
	}
	if r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseLineNullValue(t *testing.T) {
	line := `- hungry	hippo`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	// t.Errorf("%+v", r)
	if !r[0].IsNull || r[0].Value != "" {
		t.Error("expected field 1 to be null but got", r[0])
	}
	if r[1].Value != `hungry` {
		t.Error(`expected field 2 to be hungry but got`, r[1].Value)
	}
	if r[2].Value != `hippo` {
		t.Error(`expected field 3 to be hippo but got`, r[2].Value)
	}
}

func TestParseTabs(t *testing.T) {
	line := `john	"hungry ""doe"""	#this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseWithHashSignInData(t *testing.T) {
	line := `john "hungry# ""doe"""`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `hungry# "doe"` {
		t.Error(`Expect hungry "doe" but got`, r[1].Value)
	}
}

func TestParseLineWithDoubleQuotes(t *testing.T) {
	line := `john "hungry ""doe"""`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `hungry "doe"` {
		t.Error(`Expect hungry "doe" but got`, r[1].Value)
	}
}

func TestParseLineWithStartingDoubleQuotes(t *testing.T) {
	line := `john """hungry"""`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].Value != "john" {
		t.Error("Expect john but got", r[0].Value)
	}
	if r[1].Value != `"hungry"` {
		t.Error(`Expect "hungry" but got`, r)
	}
}

func TestParseLineWithoutDoubleQuotes(t *testing.T) {
	line := `john hungry doe`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
	}
}

func TestParseInvalidLine(t *testing.T) {
	line := `john hungry "doe`
	_, err := wsv.ParseLine(1, []byte(line))
	if err.Error() != `parse error on line 1, column 12 [hungry "do]: bare " in non-quoted-field` {
		t.Error(err)
	}
}

func TestParseComplexDoubleQuoteLine(t *testing.T) {
	line := `"""fName""s" """lName""" """dob""" """gender"""`
	fields, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 4 {
		t.Errorf("Parse line expected 4 record but got %+v (%d)", fields, len(fields))
		return
	}
}

func TestParseEmptyComment(t *testing.T) {
	line := `john hungry #`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Parse line expected 2 record but got %+v (%d)", r, len(r))
	}
}

func TestParseCommentWithJustSpaces(t *testing.T) {

	line := `john hungry #      `
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
	}
}

func TestReadColumnAndDataWithDoubleQuotes(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"""fName""s" """lName""" """dob""" """gender"""`)
	lines = append(lines, `john "hungry ""doe""" 2023-01-01 M`)
	lines = append(lines, `jane smith 2024-01-01 F`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	if !line.Fields[0].IsHeader || line.Fields[0].Value != `"fName"s` {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != `"lName"` {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != `"dob"` {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != `"fName"s` || line.Fields[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != `"lName"` || line.Fields[1].Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != `"dob"` || line.Fields[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != `"gender"` || line.Fields[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != `"fName"s` || line.Fields[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != `"lName"` || line.Fields[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != `"dob"` || line.Fields[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line.Fields[2])
	}
	if line.Fields[3].IsHeader || line.Fields[3].FieldName != `"gender"` || line.Fields[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line.Fields[3])
	}
}

func TestNullRemainingColumns(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"""fName""s" """lName""" """dob""" """gender"""`)
	lines = append(lines, `john "hungry ""doe""" 2023-01-01`)
	lines = append(lines, `jane smith 2024-01-01`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if !line.Fields[0].IsHeader || line.Fields[0].Value != `"fName"s` {
		t.Errorf("Header 1 does not match %+v", line.Fields[0])
	}
	if !line.Fields[1].IsHeader || line.Fields[1].Value != `"lName"` {
		t.Errorf("Header 2 does not match %+v", line.Fields[1])
	}
	if !line.Fields[2].IsHeader || line.Fields[2].Value != `"dob"` {
		t.Errorf("Header 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsHeader || line.Fields[3].Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != `"fName"s` || line.Fields[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != `"lName"` || line.Fields[1].Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != `"dob"` || line.Fields[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsNull || line.Fields[3].FieldName != `"gender"` || line.Fields[3].Value != "" {
		t.Errorf("Row 1 Column 4 does not match %+v", line.Fields[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line.Fields[0].IsHeader || line.Fields[0].FieldName != `"fName"s` || line.Fields[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line.Fields[0])
	}
	if line.Fields[1].IsHeader || line.Fields[1].FieldName != `"lName"` || line.Fields[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line.Fields[1])
	}
	if line.Fields[2].IsHeader || line.Fields[2].FieldName != `"dob"` || line.Fields[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line.Fields[2])
	}
	if !line.Fields[3].IsNull || line.Fields[3].FieldName != `"gender"` || line.Fields[3].Value != "" {
		t.Errorf("Row 2 Column 4 does not match %+v", line.Fields[3])
	}
}

func TestParseLineStartsWithNewLineAndEscapedDoubleQuotes(t *testing.T) {
	line := `john ""/"hungry"""/"""doe"/"" #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {

		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != string('\n')+`hungry"`+string('\n')+`"doe`+string('\n') {
		t.Errorf(`expected field 2 to be \\nhungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAtEndOfValueAndEscapedDoubleQuotes(t *testing.T) {
	line := `john "hungry"""/"""doe"/"" #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {

		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != `hungry"`+string('\n')+`"doe`+string('\n') {
		t.Errorf(`expected field 2 to be hungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestParseLineWithNewLineAtEndOfValueFollowedByASpaceAndEscapedDoubleQuotes(t *testing.T) {
	line := `john " hungry"""/"""doe"/" " #this is a valid comment`
	r, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[1].IsNull || r[1].Value != ` hungry"`+string('\n')+`"doe`+string('\n')+` ` {
		t.Errorf(`expected field 2 to be hungry"\\n"doe\\n but got %+v instead`, r[1])
	}
	if !r[2].IsComment {
		t.Errorf("Parse line expected the 3 field to be a comment but got %+v instead", r[2])
	}
}

func TestNewLineParsing(t *testing.T) {
	line := `"john"       "doe"       "favorite colors:"/"red"/"blue"/"green"`
	r, err := wsv.ParseLine(1, []byte(line))
	exp := "favorite colors:" + string('\n') + "red" + string('\n') + "blue" + string('\n') + "green"
	if err != nil {
		t.Error(err)
	}
	if len(r) != 3 {
		t.Errorf("Parse line expected 3 record but got %+v (%d)", r, len(r))
		return
	}
	if r[0].IsNull || r[0].Value != "john" {
		t.Errorf(`expected field 2 to be [john] but got %+v instead`, r[1])
	}
	if r[1].IsNull || r[1].Value != "doe" {
		t.Errorf(`expected field 2 to be [doe] but got %+v instead`, r[1])
	}
	if r[2].IsNull || r[2].Value != exp {
		t.Errorf("Parse line expected the 3 field to be a [%s] but got [%s] instead", exp, r[2].Value)
	}
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func TestReadDataWithNewLineInValue(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/data-with-new-line-in-value.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 3 {
		t.Error("expected to have 3 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "first name" {
		t.Error("expect header 1 to be [first name] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "last name" {
		t.Error("expect header 2 to be [last name] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "comment" {
		t.Error("expect header 3 to be [comment] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "john" {
		t.Error("expect line", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect line", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "favorite colors:\nred\nblue\ngreen" {
		t.Error("expect line", r.CurrentRow(), "field 3 to be [favorite colors:\\nred\\nblue\\ngreen] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "jane" {
		t.Error("expect line", r.CurrentRow(), "field 1 to be [jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect line", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "favorite colors:\ngreen\nyellow" {
		t.Error("expect line", r.CurrentRow(), "field 3 to be [favorite colors:\\ngreen\\nyellow] but got", line.Fields[2].Value, "instead")
	}
}

func TestReadDataWithSpaces(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/data-with-spaces.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Given Name" {
		t.Error("expect header 1 to be [Given Name] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "Family Name" {
		t.Error("expect header 2 to be [Family Name] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Date of Birth" {
		t.Error("expect header 3 to be [Date of Birth] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "Favorite Color" {
		t.Error("expect header 4 to be [Favorite Color] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Jean Smith" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Jean Smith] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "Le Croix" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Le Croix] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Jan 01 2023" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Jan 01 2023] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "Space Purple" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Space Purple] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Mary Jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Mary Jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "Vasquez Rojas" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Vasquez Rojas] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Feb 02 2021" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Feb 02 2021] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "Midnight Grey" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Midnight Grey] but got", line.Fields[3].Value, "instead")
	}
}

func TestReadDataWithoutSpace(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/data-without-spaces.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "jane" {
		t.Error("expect row 3, field 1 to be [jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "smith" {
		t.Error("expect row 3, field 2 to be [smith] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2024-01-01" {
		t.Error("expect row 3, field 3 to be [2024-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "F" {
		t.Error("expect row 3, field 4 to be [F] but got", line.Fields[3].Value, "instead")
	}
}

func TestReadOmittedColumnsToNull(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/omitted-columns-to-null.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", line.Fields[2].Value, "instead")
	}
	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "jane" {
		t.Error("expect row 3, field 1 to be [jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "smith" {
		t.Error("expect row 3, field 2 to be [smith] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", line.Fields[2].Value, "instead")
	}
	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", line.Fields[3].Value, "instead")
	}
}

func TestReadSimpleWithTabs(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/simple-with-tabs.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 3 {
		t.Error("expected to have 3 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Name" {
		t.Error("expect header 1 to be [Name] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "Age" {
		t.Error("expect header 2 to be [Age] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Color" {
		t.Error("expect header 3 to be [Color] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Scott" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Scott] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "21" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [21] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Red" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Red] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Josh" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Josh] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "18" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [18] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Blue" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Blue] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "34" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [34] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Yellow" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Yellow] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Bob" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Bob] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "16" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [16] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Pink" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Pink] but got", line.Fields[2].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if line.Fields[0].Value != "Ashley" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Ashley] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "47" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [47] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Red" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Red] but got", line.Fields[2].Value, "instead")
	}
}

func TestReadWithComments(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/with-comments.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", line.Fields[3].Value, "instead")
	}
	if line.Comment != "these are headers" {
		t.Error("expect comment in header to be [these are headers] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}
	if line.Fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", line.Fields[3].Value, "instead")
	}
	if line.Comment != "this data is probably not accurate" {
		t.Error("expect row", r.CurrentRow(), "to have a comment be [this data is probably not accurate] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}
	if line.Fields[0].Value != "jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "smith" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [smith] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected to have 0 data fields but got", len(line.Fields), "instead")
		return
	}
	if line.Comment != "this is a comment" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", line.Comment, "instead")
	}

}

func TestReadWithCEmptyLinesAndComments(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/with-empty-lines-and-comments.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", line.Fields[3].Value, "instead")
	}
	if line.Comment != "these are headers" {
		t.Error("expect header 5 to be [these are headers] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", len(line.Fields), "instead")
		return
	}
	if line.Comment != "here we go!" {
		t.Error("expect row", r.CurrentRow(), "field ` to be [here we go!] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", line.Fields[3].Value, "instead")
	}
	if line.Comment != "this data is probably not accurate" {
		t.Error("expect row", r.CurrentRow(), "field 5 to be [this data is probably not accurate] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected to have 4 fields but got", len(line.Fields), "instead")
		return
	}
	if line.Fields[0].Value != "jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "smith" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [smith] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected to have 0 data fields but got", len(line.Fields), "instead")
		return
	}
	if line.Comment != "this is a comment" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [this is a comment] but got", line.Comment, "instead")
	}

}

func TestReadInvalidCommentPlacementDueToOmittedFields(t *testing.T) {
	file, err := os.Open(fmt.Sprintf("%s/examples/invalid-comment-placement-due-to-omitted-fields.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "name" {
		t.Error("expect header 1 to be [name] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "jersey" {
		t.Error("expect header 2 to be [jersey] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "team" {
		t.Error("expect header 3 to be [team] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "sport" {
		t.Error("expect header 4 to be [sport] but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "15" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [15] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "tigers" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [tigers] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "baseball" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [baseball] but got", line.Fields[3].Value, "instead")
	}
	line, err = r.Read()
	if err == nil {
		t.Error("expected to return an error but did not", line)
		return
	}

}

func TestReadComplexValues(t *testing.T) {
	file, err := os.Open(fmt.Sprintf("%s/examples/complex-values.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	line, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Country" {
		t.Error("expect header 1 to be [Country] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[1].Value != "Capital" {
		t.Error("expect header 2 to be [Capital] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[2].Value != "Emoji of Flag" {
		t.Error("expect header 3 to be [Emoji of Flag] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[3].Value != "Interesting Facts" {
		t.Error("expect header 4 to be [Interesting Facts] but got", line.Fields[3].Value, "instead")
	}
	if line.Comment != "facts generated from Google's Gemini 2024-04-24" {
		t.Error("expected header 5 to be comment and have the value [facts generated from Google's Gemini 2024-04-24] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "France" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [France] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Paris" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Paris] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇫🇷" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇫🇷] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "The Eiffel Tower was built for the 1889 World's Fair."+string('\n')+"It was almost torn down afterwards." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Eiffel Tower was built for the 1889 World's Fair.\\nIt was almost torn down afterwards.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Germany" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Germany] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Berlin" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Berlin] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇩🇪" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇩🇪] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "Germany has over 2,000 beer breweries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Germany has over 2,000 beer breweries.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Italy" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Italy] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Rome" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Rome] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇮🇹" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇮🇹] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "The Colosseum in Rome could hold an estimated 50,000 spectators." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Colosseum in Rome could hold an estimated 50,000 spectators.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Japan" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Japan] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Tokyo" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Tokyo] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇯🇵🇯🇵" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇯🇵🇯🇵] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "Japan is a volcanic archipelago with over 100 active volcanoes."+string('\n')+"The currency is the yen and the symbol is ¥." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Japan is a volcanic archipelago with over 100 active volcanoes.\\nThe currency is the yen and the symbol is ¥.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	if line.Comment != "has half-width characters" {
		t.Error("expect row", r.CurrentRow(), "field 5 to be a comment but got", line.Fields[3].Value, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Spain" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Spain] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Madrid" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Madrid] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇪🇸" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇪🇸] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "Spain has the second highest number of UNESCO World Heritage Sites in the world." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Spain has the second highest number of UNESCO World Heritage Sites in the world.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "United Kingdom" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [United Kingdom] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "London" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [London] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇬🇧" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇬🇧] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "The United Kingdom is a parliamentary monarchy with a rich history dating back centuries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United Kingdom is a parliamentary monarchy with a rich history dating back centuries.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Comment != " emphasis on 50 with double quotes" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [ emphasis on 50 with double quotes] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead", line)
		return
	}

	if line.Fields[0].Value != "United States of America" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [United States of America] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Washington D.C." {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Washington D.C.] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇺🇸 🏴‍☠️" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇺🇸 🏴‍☠️] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if line.Fields[3].Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", line.Fields[3].FieldName, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Comment != " update the remaining" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [ update the remaining] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "India" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [India] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇮🇳" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇮🇳] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Errorf("expect row %d field 4 to have value [NULL] but got %+v instead", r.CurrentRow(), line.Fields[3])
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 data fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Canada" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Canada] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Ottawa" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Ottawa] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	if line.Comment != "need to add facts for the remaining" {
		t.Error("expect row", r.CurrentRow(), "field 5 to have value [need to add facts for the remaining] but got", line.Comment, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Australia" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Australia] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Canberra" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Canberra] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇦🇺" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇦🇺] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Brazil" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Brazil] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Brasília" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Brasília] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇧🇷" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇧🇷] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Argentina" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Argentina] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Buenos Aires" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Buenos Aires] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇦🇷" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇦🇷] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Mexico" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Mexico] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Mexico City" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Mexico City] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇲🇽" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇲🇽] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "China" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [China] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Beijing" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Beijing] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇨🇳" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇨🇳] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(line.Fields), "instead")
		return
	}

	if line.Fields[0].Value != "Russia" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Russia] but got", line.Fields[0].Value, "instead")
	}
	if line.Fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", line.Fields[0].FieldName, "instead")
	}

	if line.Fields[1].Value != "Moscow" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Moscow] but got", line.Fields[1].Value, "instead")
	}
	if line.Fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", line.Fields[1].FieldName, "instead")
	}

	if line.Fields[2].Value != "🇷🇺" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇷🇺] but got", line.Fields[2].Value, "instead")
	}
	if line.Fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", line.Fields[2].FieldName, "instead")
	}

	if !line.Fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", line.Fields[3].Value, "instead")
	}
	if line.Fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", line.Fields[3].FieldName, "instead")
	}

	line, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(line.Fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(line.Fields), "instead")
		return
	}
	_, err = r.Read()
	if err != io.EOF {
		t.Error("expected an error EOF")
	}

	_, err = r.Read()
	if err != wsv.ErrReaderEnded {
		t.Error("expected an error ErrReaderEnded")
	}
}

func TestParseLineTrailingWhiteSpace(t *testing.T) {
	line := `Mexico						"Mexico City"		🇲🇽			  -	`
	fields, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Errorf("expect to have 4 fields but got %+v instead", fields)
		return
	}

	if fields[0].Value != "Mexico" {
		t.Error("field 1 to be [Mexico] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Mexico City" {
		t.Error("field 2 to be [Mexico City] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "🇲🇽" {
		t.Errorf("expect row %d field 3 to have value [NULL] but got %+v instead", 1, fields[2])
	}
	if !fields[3].IsNull {
		t.Error("field 4 to have value [NULL] but got", fields[3].Value, "instead")
	}
}

func TestParseLineWithEmojisAndEscapedDoubleQuotesSurroundedByWhitespace(t *testing.T) {
	line := `"United States of America"  "Washington D.C." 	"🇺🇸 🏴‍☠️"           "The United States of America is a federal republic with ""50"" states."`
	fields, err := wsv.ParseLine(1, []byte(line))
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line to have 4 fields but got", len(fields), "instead", fields)
		return
	}

	if fields[0].Value != "United States of America" {
		t.Error("field 1 to be [United States of America] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Washington D.C." {
		t.Error("field 2 to be [Washington D.C.] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "🇺🇸 🏴‍☠️" {
		t.Error("field 3 to be [🇺🇸 🏴‍☠️] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", fields[3].Value, "instead")
	}
}

func TestReaderToDocument(t *testing.T) {
	file, err := os.Open(fmt.Sprintf("%s/examples/complex-values.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	doc, err := r.ToDocument()
	if err != nil {
		t.Error(err)
		return
	}
	if doc.LineCount() != 30 {
		t.Error("Line count is", doc.LineCount())
		return
	}
	line, err := doc.Line(29)
	if err != nil {
		t.Error(err)
		return
	}
	v, err := line.NextField()
	if err != nil {
		t.Error(err)
		return
	}
	if v.Value != "Russia" {
		t.Error("expected the value to be Russia but got", v.Value, "instead")
	}
	line, err = doc.AddLine()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("South Korea")
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Seoul")
	if err != nil {
		t.Error(err)
		return
	}

	err = line.AppendNull()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!")
	if err != nil {
		t.Error(err)
		return
	}
	line.Comment = "added via document writer"

	data, err := doc.WriteAll()
	if err != nil {
		t.Error(err)
		return
	}
	file, err = os.Create(fmt.Sprintf("%s/example-output/complex-output.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		t.Error(err)
	}

}
