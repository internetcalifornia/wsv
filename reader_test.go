package wsv_test

import (
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

func TestReadColumnAndDataWithSpaces(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"Given Name" "Family Name" "Date of Birth" "Favorite Color"`)
	lines = append(lines, `"Jean Smith" "Le Croix" "Jan 01 2023" "Space Purple"`)
	lines = append(lines, `"Mary Jane" "Vasquez Rojas" "Feb 02 2021" "Midnight Grey"`)
	file := strings.Join(lines, string('\n'))
	str := strings.NewReader(file)
	r := wsv.NewReader(str)
	fields, err := r.Read()
	r.ReadAll()
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
	line := `john ""hungry""`
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

func TestReadColumnAndDataWithDoubleQuotes(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `""fName""s ""lName"" ""dob"" ""gender""`)
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
	lines = append(lines, `""fName""s ""lName"" ""dob"" ""gender""`)
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
