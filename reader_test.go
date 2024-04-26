package wsv_test

import (
	"fmt"
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
	fields, err := r.Read()

	if err != nil {
		t.Error(err)
	}
	if len(fields) < 4 {
		t.Errorf("expected 4 fields but only got %d instead", len(fields))
	}
	if !fields[0].IsHeader || fields[0].Value != "Given Name" {
		t.Errorf("Header 1 does not match %+v", fields[0])
	}
	if !fields[1].IsHeader || fields[1].Value != "Family Name" {
		t.Errorf("Header 2 does not match %+v", fields[1])
	}
	if !fields[2].IsHeader || fields[2].Value != "Date of Birth" {
		t.Errorf("Header 3 does not match %+v", fields[2])
	}
	if !fields[3].IsHeader || fields[3].Value != "Favorite Color" {
		t.Errorf("Header 4 does not match %+v", fields[3])
	}
	fields, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if fields[0].IsHeader || fields[0].FieldName != "Given Name" || fields[0].Value != "Jean Smith" {
		t.Errorf("Row 1 Column 1 does not match %+v", fields[0])
	}
	if fields[1].IsHeader || fields[1].FieldName != "Family Name" || fields[1].Value != "Le Croix" {
		t.Errorf("Row 1 Column 2 does not match %+v", fields[1])
	}
	if fields[2].IsHeader || fields[2].FieldName != "Date of Birth" || fields[2].Value != "Jan 01 2023" {
		t.Errorf("Row 1 Column 3 does not match %+v", fields[2])
	}
	if fields[3].IsHeader || fields[3].FieldName != "Favorite Color" || fields[3].Value != "Space Purple" {
		t.Errorf("Row 1 Column 4 does not match %+v", fields[3])
	}
	fields, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if fields[0].IsHeader || fields[0].FieldName != "Given Name" || fields[0].Value != "Mary Jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", fields[0])
	}
	if fields[1].IsHeader || fields[1].FieldName != "Family Name" || fields[1].Value != "Vasquez Rojas" {
		t.Errorf("Row 2 Column 2 does not match %+v", fields[1])
	}
	if fields[2].IsHeader || fields[2].FieldName != "Date of Birth" || fields[2].Value != "Feb 02 2021" {
		t.Errorf("Row 2 Column 3 does not match %+v", fields[2])
	}
	if fields[3].IsHeader || fields[3].FieldName != "Favorite Color" || fields[3].Value != "Midnight Grey" {
		t.Errorf("Row 2 Column 4 does not match %+v", fields[3])
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
	if !line[0].IsHeader || line[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line[0])
	}
	if !line[1].IsHeader || line[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line[1])
	}
	if !line[2].IsHeader || line[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line[2])
	}
	if !line[3].IsHeader || line[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line[3])
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
	if !line[0].IsHeader || line[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line[0])
	}
	if !line[1].IsHeader || line[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line[1])
	}
	if !line[2].IsHeader || line[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line[2])
	}
	if !line[3].IsHeader || line[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(r.Comments) != 2 {
		t.Errorf("Expected to have three comments but got %d %+v", len(r.Comments), r.Comments)
	}
	if c, err := r.CommentFor(2); err != nil || c != "#this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %d %+v", len(r.Comments), err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "doe" {
		t.Errorf("Row 1 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if len(line) != 1 {
		t.Error("expected the line to have a length of 1 but got", len(line), "instead")
		return
	}

	if !line[0].IsComment {
		t.Error("expected this line to contain a comment but got", line[0], "instead")
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
	if !line[0].IsHeader || line[0].Value != "fName" {
		t.Errorf("Header 1 does not match %+v", line[0])
	}
	if !line[1].IsHeader || line[1].Value != "lName" {
		t.Errorf("Header 2 does not match %+v", line[1])
	}
	if !line[2].IsHeader || line[2].Value != "dob" {
		t.Errorf("Header 3 does not match %+v", line[2])
	}
	if !line[3].IsHeader || line[3].Value != "gender" {
		t.Errorf("Header 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line) != 0 {
		t.Error("expected row 2 to have zero fields but got", len(line), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line) != 0 {
		t.Error("expected row 3 to have zero fields but got", len(line), "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(line) != 1 {
		t.Error("expected row 4 to have 1 comment but got", len(line), "instead")
		return
	}
	if !line[0].IsComment {
		t.Error("expected a comment but got", line[0], "instead")
		return
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if len(r.Comments) != 3 {
		t.Errorf("Expected parse until row 5 to have three comments but got %d %+v", len(r.Comments), r.Comments)
	}
	if c, err := r.CommentFor(5); err != nil || c != "#this data is probably not accurate" {
		t.Errorf("Expected row to have a comment but got %d %+v", len(r.Comments), err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "john" {
		t.Errorf("Row 5 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "doe" {
		t.Errorf("Row 5 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2023-01-01" {
		t.Errorf("Row 5 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "M" {
		t.Errorf("Row 5 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != "fName" || line[0].Value != "jane" {
		t.Errorf("Row 6 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != "lName" || line[1].Value != "smith" {
		t.Errorf("Row 6 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != "dob" || line[2].Value != "2024-01-01" {
		t.Errorf("Row 6 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != "gender" || line[3].Value != "F" {
		t.Errorf("Row 6 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	if len(line) != 1 {
		t.Error("expected row 7 to have a length of 1 but got", len(line), "instead")
		return
	}

	if !line[0].IsComment {
		t.Error("expected row 7 to contain a comment but got", line[0], "instead")
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

	if !line[0].IsHeader || line[0].Value != `"fName"s` {
		t.Errorf("Header 1 does not match %+v", line[0])
	}
	if !line[1].IsHeader || line[1].Value != `"lName"` {
		t.Errorf("Header 2 does not match %+v", line[1])
	}
	if !line[2].IsHeader || line[2].Value != `"dob"` {
		t.Errorf("Header 3 does not match %+v", line[2])
	}
	if !line[3].IsHeader || line[3].Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != `"fName"s` || line[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != `"lName"` || line[1].Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != `"dob"` || line[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != `"gender"` || line[3].Value != "M" {
		t.Errorf("Row 1 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != `"fName"s` || line[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != `"lName"` || line[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != `"dob"` || line[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line[2])
	}
	if line[3].IsHeader || line[3].FieldName != `"gender"` || line[3].Value != "F" {
		t.Errorf("Row 2 Column 4 does not match %+v", line[3])
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
	if !line[0].IsHeader || line[0].Value != `"fName"s` {
		t.Errorf("Header 1 does not match %+v", line[0])
	}
	if !line[1].IsHeader || line[1].Value != `"lName"` {
		t.Errorf("Header 2 does not match %+v", line[1])
	}
	if !line[2].IsHeader || line[2].Value != `"dob"` {
		t.Errorf("Header 3 does not match %+v", line[2])
	}
	if !line[3].IsHeader || line[3].Value != `"gender"` {
		t.Errorf("Header 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != `"fName"s` || line[0].Value != "john" {
		t.Errorf("Row 1 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != `"lName"` || line[1].Value != `hungry "doe"` {
		t.Errorf("Row 1 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != `"dob"` || line[2].Value != "2023-01-01" {
		t.Errorf("Row 1 Column 3 does not match %+v", line[2])
	}
	if !line[3].IsNull || line[3].FieldName != `"gender"` || line[3].Value != "" {
		t.Errorf("Row 1 Column 4 does not match %+v", line[3])
	}
	line, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if line[0].IsHeader || line[0].FieldName != `"fName"s` || line[0].Value != "jane" {
		t.Errorf("Row 2 Column 1 does not match %+v", line[0])
	}
	if line[1].IsHeader || line[1].FieldName != `"lName"` || line[1].Value != "smith" {
		t.Errorf("Row 2 Column 2 does not match %+v", line[1])
	}
	if line[2].IsHeader || line[2].FieldName != `"dob"` || line[2].Value != "2024-01-01" {
		t.Errorf("Row 2 Column 3 does not match %+v", line[2])
	}
	if !line[3].IsNull || line[3].FieldName != `"gender"` || line[3].Value != "" {
		t.Errorf("Row 2 Column 4 does not match %+v", line[3])
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
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 3 {
		t.Error("expected to have 3 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "first name" {
		t.Error("expect header 1 to be [first name] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "last name" {
		t.Error("expect header 2 to be [last name] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "comment" {
		t.Error("expect header 3 to be [comment] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "john" {
		t.Error("expect line", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect line", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "favorite colors:\nred\nblue\ngreen" {
		t.Error("expect line", r.CurrentRow(), "field 3 to be [favorite colors:\\nred\\nblue\\ngreen] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "jane" {
		t.Error("expect line", r.CurrentRow(), "field 1 to be [jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect line", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "favorite colors:\ngreen\nyellow" {
		t.Error("expect line", r.CurrentRow(), "field 3 to be [favorite colors:\\ngreen\\nyellow] but got", fields[2].Value, "instead")
	}
}

func TestReadDataWithSpaces(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/data-with-spaces.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Given Name" {
		t.Error("expect header 1 to be [Given Name] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Family Name" {
		t.Error("expect header 2 to be [Family Name] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Date of Birth" {
		t.Error("expect header 3 to be [Date of Birth] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "Favorite Color" {
		t.Error("expect header 4 to be [Favorite Color] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Jean Smith" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Jean Smith] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Le Croix" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Le Croix] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Jan 01 2023" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Jan 01 2023] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "Space Purple" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Space Purple] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Mary Jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Mary Jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Vasquez Rojas" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Vasquez Rojas] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Feb 02 2021" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Feb 02 2021] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "Midnight Grey" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [Midnight Grey] but got", fields[3].Value, "instead")
	}
}

func TestReadDataWithoutSpace(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/data-without-spaces.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "jane" {
		t.Error("expect row 3, field 1 to be [jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "smith" {
		t.Error("expect row 3, field 2 to be [smith] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2024-01-01" {
		t.Error("expect row 3, field 3 to be [2024-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "F" {
		t.Error("expect row 3, field 4 to be [F] but got", fields[3].Value, "instead")
	}
}

func TestReadOmittedColumnsToNull(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/omitted-columns-to-null.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", fields[2].Value, "instead")
	}
	if !fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "jane" {
		t.Error("expect row 3, field 1 to be [jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "smith" {
		t.Error("expect row 3, field 2 to be [smith] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", fields[2].Value, "instead")
	}
	if !fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [NULL] but got", fields[3].Value, "instead")
	}
}

func TestReadSimpleWithTabs(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/simple-with-tabs.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 3 {
		t.Error("expected to have 3 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Name" {
		t.Error("expect header 1 to be [Name] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Age" {
		t.Error("expect header 2 to be [Age] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Color" {
		t.Error("expect header 3 to be [Color] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Scott" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Scott] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "21" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [21] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Red" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Red] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Josh" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Josh] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "18" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [18] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Blue" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Blue] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "34" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [34] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Yellow" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Yellow] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Bob" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Bob] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "16" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [16] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Pink" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Pink] but got", fields[2].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if fields[0].Value != "Ashley" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Ashley] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "47" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [47] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Red" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [Red] but got", fields[2].Value, "instead")
	}
}

func TestReadWithComments(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/with-comments.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected to have 5 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", fields[3].Value, "instead")
	}
	if fields[4].Value != "#these are headers" || !fields[4].IsComment {
		t.Error("expect header 5 to be [#these are headers] but got", fields[4].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected to have 5 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", fields[3].Value, "instead")
	}
	if fields[4].Value != "#this data is probably not accurate" || !fields[4].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 5 to be [#this data is probably not accurate] but got", fields[4].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected to have 4 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "smith" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [smith] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 1 {
		t.Error("expected to have 1 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "#this is a comment" || !fields[0].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", fields[0].Value, "instead")
	}

}

func TestReadWithCEmptyLinesAndComments(t *testing.T) {

	file, err := os.Open(fmt.Sprintf("%s/examples/with-empty-lines-and-comments.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected line", r.CurrentRow(), "to have 5 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "fName" {
		t.Error("expect header 1 to be [fName] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "lName" {
		t.Error("expect header 2 to be [lName] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "dob" {
		t.Error("expect header 3 to be [dob] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "gender" {
		t.Error("expect header 4 to be [gender] but got", fields[3].Value, "instead")
	}
	if fields[4].Value != "#these are headers" || !fields[4].IsComment {
		t.Error("expect header 5 to be [#these are headers] but got", fields[4].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}
	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}
	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 1 {
		t.Error("expected line", r.CurrentRow(), "to have 1 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "#here we go!" || !fields[0].IsComment {
		t.Error("expect row", r.CurrentRow(), "field ` to be [#here we go!] but got", fields[0].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected line", r.CurrentRow(), "to have 5 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "doe" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [doe] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2023-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2023-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "M" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [M] but got", fields[3].Value, "instead")
	}
	if fields[4].Value != "#this data is probably not accurate" || !fields[4].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 5 to be [#this data is probably not accurate] but got", fields[4].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected to have 4 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "jane" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "smith" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [smith] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "2024-01-01" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [2024-01-01] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "F" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [F] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 1 {
		t.Error("expected to have 1 fields but got", len(fields), "instead")
		return
	}
	if fields[0].Value != "#this is a comment" || !fields[0].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [jane] but got", fields[0].Value, "instead")
	}

}

func TestReadInvalidCommentPlacementDueToOmittedFields(t *testing.T) {
	file, err := os.Open(fmt.Sprintf("%s/examples/invalid-comment-placement-due-to-omitted-fields.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	r := wsv.NewReader(file)
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "name" {
		t.Error("expect header 1 to be [name] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "jersey" {
		t.Error("expect header 2 to be [jersey] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "team" {
		t.Error("expect header 3 to be [team] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "sport" {
		t.Error("expect header 4 to be [sport] but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "john" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [john] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "15" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [15] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "tigers" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [tigers] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "baseball" {
		t.Error("expect row", r.CurrentRow(), "field 4 to be [baseball] but got", fields[3].Value, "instead")
	}
	fields, err = r.Read()
	if err == nil {
		t.Error("expected to return an error but did not", fields)
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
	fields, err := r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected line", r.CurrentRow(), "to have 5 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Country" {
		t.Error("expect header 1 to be [Country] but got", fields[0].Value, "instead")
	}
	if fields[1].Value != "Capital" {
		t.Error("expect header 2 to be [Capital] but got", fields[1].Value, "instead")
	}
	if fields[2].Value != "Emoji of Flag" {
		t.Error("expect header 3 to be [Emoji of Flag] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "Interesting Facts" {
		t.Error("expect header 4 to be [Interesting Facts] but got", fields[3].Value, "instead")
	}
	if fields[4].Value != "#facts generated from Google's Gemini 2024-04-24" || !fields[4].IsComment {
		t.Error("expected header 5 to be comment and have the value [#facts generated from Google's Gemini 2024-04-24] but got", fields[4].Value, fields[4].IsComment, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "France" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [France] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Paris" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Paris] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇫🇷" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇫🇷] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "The Eiffel Tower was built for the 1889 World's Fair."+string('\n')+"It was almost torn down afterwards." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Eiffel Tower was built for the 1889 World's Fair.\\nIt was almost torn down afterwards.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Germany" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Germany] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Berlin" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Berlin] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇩🇪" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇩🇪] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "Germany has over 2,000 beer breweries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Germany has over 2,000 beer breweries.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Italy" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Italy] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Rome" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Rome] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇮🇹" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇮🇹] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "The Colosseum in Rome could hold an estimated 50,000 spectators." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The Colosseum in Rome could hold an estimated 50,000 spectators.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 5 {
		t.Error("expected line", r.CurrentRow(), "to have 5 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Japan" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Japan] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Tokyo" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Tokyo] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇯🇵🇯🇵" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇯🇵🇯🇵] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "Japan is a volcanic archipelago with over 100 active volcanoes."+string('\n')+"The currency is the yen and the symbol is ¥." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Japan is a volcanic archipelago with over 100 active volcanoes.\\nThe currency is the yen and the symbol is ¥.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	if fields[4].Value != "#has half-width characters" || !fields[4].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 5 to be a comment but got", fields[3].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "Spain" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [Spain] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Madrid" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Madrid] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇪🇸" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇪🇸] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "Spain has the second highest number of UNESCO World Heritage Sites in the world." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [Spain has the second highest number of UNESCO World Heritage Sites in the world.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "United Kingdom" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [United Kingdom] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "London" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [London] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇬🇧" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇬🇧] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "The United Kingdom is a parliamentary monarchy with a rich history dating back centuries." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United Kingdom is a parliamentary monarchy with a rich history dating back centuries.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 1 {
		t.Error("expected line", r.CurrentRow(), "to have 1 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "# emphasis on 50 with double quotes" || !fields[0].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [# emphasis on 50 with double quotes] but got", fields[0].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead", fields)
		return
	}

	if fields[0].Value != "United States of America" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [United States of America] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "Washington D.C." {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [Washington D.C.] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇺🇸 🏴‍☠️" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇺🇸 🏴‍☠️] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 0 {
		t.Error("expected line", r.CurrentRow(), "to have 0 fields but got", len(fields), "instead")
		return
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 1 {
		t.Error("expected line", r.CurrentRow(), "to have 1 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "# update the remaining" || !fields[0].IsComment {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [# update the remaining] but got", fields[0].Value, "instead")
	}

	fields, err = r.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if len(fields) != 4 {
		t.Error("expected line", r.CurrentRow(), "to have 4 fields but got", len(fields), "instead")
		return
	}

	if fields[0].Value != "India" {
		t.Error("expect row", r.CurrentRow(), "field 1 to be [India] but got", fields[0].Value, "instead")
	}
	if fields[0].FieldName != "Country" {
		t.Error("expect row", r.CurrentRow(), "field 1 to have field name [Country] but got", fields[0].FieldName, "instead")
	}

	if fields[1].Value != "" {
		t.Error("expect row", r.CurrentRow(), "field 2 to be [] but got", fields[1].Value, "instead")
	}
	if fields[1].FieldName != "Capital" {
		t.Error("expect row", r.CurrentRow(), "field 2 to have field name [Capital] but got", fields[1].FieldName, "instead")
	}

	if fields[2].Value != "🇮🇳" {
		t.Error("expect row", r.CurrentRow(), "field 3 to be [🇮🇳] but got", fields[2].Value, "instead")
	}
	if fields[2].FieldName != "Emoji of Flag" {
		t.Error("expect row", r.CurrentRow(), "field 3 to have field name [Emoji of Flag] but got", fields[2].FieldName, "instead")
	}

	if fields[3].IsNull {
		t.Error("expect row", r.CurrentRow(), "field 4 to have value [NULL] but got", fields[3].FieldName, "instead")
	}
	if fields[3].FieldName != "Interesting Facts" {
		t.Error("expect row", r.CurrentRow(), "field 4 to have field name [Interesting Facts] but got", fields[3].FieldName, "instead")
	}

}

func TestParseLineWithEmojisAndEscapedDoubleQuotesSurroundedByWhitespace(t *testing.T) {
	line := `"United States of America"  "Washington D.C." 	"flag"           "The United States of America is a federal republic with ""50"" states."`
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
	if fields[2].Value != "flag" {
		t.Error("field 3 to be [flag] but got", fields[2].Value, "instead")
	}
	if fields[3].Value != "The United States of America is a federal republic with \"50\" states." {
		t.Error("field 4 to have value [The United States of America is a federal republic with \"50\" states.] but got", fields[3].Value, "instead")
	}
}
