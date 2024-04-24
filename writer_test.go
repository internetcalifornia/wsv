package wsv_test

import (
	"testing"

	"github.com/internetcalifornia/wsv"
)

func TestSerializeSimpleText(t *testing.T) {
	r := wsv.RecordField{Value: "Age"}
	txt := r.SerializeText()
	if txt != "Age" {
		t.Error("expected [Age] but got", "["+txt+"]", "instead")
	}
}

func TestSerializeNull(t *testing.T) {
	r := wsv.RecordField{Value: "", IsNull: true}
	txt := r.SerializeText()
	if txt != "-" {
		t.Error("expected [-] but got", "["+txt+"]", "instead")
	}
}

func TestSerializeWithSpaceInValue(t *testing.T) {
	r := wsv.RecordField{Value: "First Name"}
	txt := r.SerializeText()
	if txt != `"First Name"` {
		t.Error("expected [\"First Name\"] but got", "["+txt+"]", "instead")
	}
}

func TestSerializeWithNewLineInValue(t *testing.T) {
	r := wsv.RecordField{Value: "First\nName"}
	txt := r.SerializeText()
	if txt != `"First"/"Name"` {
		t.Error("expected [\"First\"/\"Name\"] but got", "["+txt+"]", "instead")
	}
}

func TestSerializeWitDoubleQuotesInValue(t *testing.T) {
	r := wsv.RecordField{Value: `Patient "Alias" or Nickname`}
	txt := r.SerializeText()
	if txt != `"Patient ""Alias"" or Nickname"` {
		t.Error(`expected ["Patient ""Alias"" or Nickname"] but got`, "["+txt+"]", "instead")
	}
}

func TestWithEmptyLines(t *testing.T) {
	w := wsv.NewWriter()

	r := [][]wsv.RecordField{
		{
			{IsNull: false, Value: "Age", IsHeader: true, FieldIndex: 0, FieldName: "Age", RowIndex: 1},
			{IsNull: false, Value: "Name", IsHeader: true, FieldIndex: 1, FieldName: "Name", RowIndex: 1},
			{IsNull: false, Value: "# these are headers", IsComment: true, FieldIndex: 2, RowIndex: 1},
		},
		{
			{IsNull: false, Value: "23", FieldName: "Age", FieldIndex: 0, RowIndex: 2},
			{IsNull: false, Value: "Bob", FieldIndex: 1, FieldName: "Name", RowIndex: 2},
		},
		{
			{Value: "", FieldIndex: 0, FieldName: "Age", RowIndex: 3},
		},
		{
			{Value: "", FieldIndex: 0, FieldName: "Age", RowIndex: 4},
		},
		{
			{Value: "50", FieldName: "Age", FieldIndex: 0, RowIndex: 5},
			{IsNull: false, Value: "Alice", FieldIndex: 1, FieldName: "Name", RowIndex: 5},
		},
	}

	for i := range r {
		for _, f := range r[i] {
			err := w.Append(i, f)
			if err != nil {
				t.Error(err)
			}
		}
	}

	if len(w.Records) != 5 {
		t.Error("expected to have 5 rows but got", len(w.Records), "instead")
		t.FailNow()
	}

	str := w.Write()
	if str == "" {
		t.Error("expected a output but got an empty string")
		t.FailNow()
	}
	exp := "Age Name # these are headers\n23  Bob\n\n\n50  Alice\n"
	if str != exp {
		t.Error("expect the format to match\n", `[`+exp+`]`, "but got\n", `[`+str+`]`, "instead")
		t.Error("Length of expected", len(exp), "length of string", len(str))
	}

	if len(w.Headers) != 2 {
		t.Error("expect to have 2 headers but got", len(w.Headers), "instead")
		return
	}

	if w.Headers[0] != "Age" {
		t.Error("expect the first header to have the value of [Age] but got", w.Headers[0], "instead")
	}

	if w.Headers[1] != "Name" {
		t.Error("expect the first header to have the value of [Name] but got", w.Headers[0], "instead")
	}
}

func TestWithComments(t *testing.T) {
	w := wsv.NewWriter()

	r := [][]wsv.RecordField{
		{
			{IsNull: false, Value: "Age", IsHeader: true, FieldIndex: 0, FieldName: "Age", RowIndex: 1},
			{IsNull: false, Value: "Name", IsHeader: true, FieldIndex: 1, FieldName: "Name", RowIndex: 1},
			{IsNull: false, Value: "# these are headers", IsComment: true, FieldIndex: 2, RowIndex: 1},
		},
		{
			{IsNull: false, Value: "23", FieldName: "Age", FieldIndex: 0, RowIndex: 2},
			{IsNull: false, Value: "Bob", FieldIndex: 1, FieldName: "Name", RowIndex: 2},
		},
		{
			{Value: "50", FieldName: "Age", FieldIndex: 0, RowIndex: 3},
			{IsNull: false, Value: "Alice", FieldIndex: 1, FieldName: "Name", RowIndex: 3},
		},
	}

	for i := range r {
		for _, f := range r[i] {
			err := w.Append(i, f)
			if err != nil {
				t.Error(err)
			}
		}
	}

	if len(w.Records) != 3 {
		t.Error("expected to have 3 rows but got", len(w.Records), "instead")
		t.FailNow()
	}

	str := w.Write()
	if str == "" {
		t.Error("expected a output but got an empty string")
		t.FailNow()
	}
	exp := "Age Name # these are headers\n23  Bob\n50  Alice\n"
	if str != exp {
		t.Error("expect the format to match\n", `[`+exp+`]`, "but got\n", `[`+str+`]`, "instead")
		t.Error("Length of expected", len(exp), "length of string", len(str))
	}

	if len(w.Headers) != 2 {
		t.Error("expect to have 2 headers but got", len(w.Headers), "instead")
		return
	}

	if w.Headers[0] != "Age" {
		t.Error("expect the first header to have the value of [Age] but got", w.Headers[0], "instead")
	}

	if w.Headers[1] != "Name" {
		t.Error("expect the first header to have the value of [Name] but got", w.Headers[0], "instead")
	}
}

func TestSimpleWriteToString(t *testing.T) {
	w := wsv.NewWriter()

	r := [][]wsv.RecordField{
		{
			{IsNull: false, Value: "Age", IsHeader: true, FieldIndex: 0, FieldName: "Age", RowIndex: 1},
			{IsNull: false, Value: "Name", IsHeader: true, FieldIndex: 1, FieldName: "Name", RowIndex: 1},
		},
		{
			{IsNull: false, Value: "23", FieldName: "Age", FieldIndex: 0, RowIndex: 2},
			{IsNull: false, Value: "Bob", FieldIndex: 1, FieldName: "Name", RowIndex: 2},
		},
		{
			{Value: "50", FieldName: "Age", FieldIndex: 0, RowIndex: 3},
			{IsNull: false, Value: "Alice", FieldIndex: 1, FieldName: "Name", RowIndex: 3},
		},
	}

	for i := range r {
		for _, f := range r[i] {
			err := w.Append(i, f)
			if err != nil {
				t.Error(err)
			}
		}
	}

	if len(w.Records) != 3 {
		t.Error("expected to have 3 rows but got", len(w.Records), "instead")
		t.FailNow()
	}

	str := w.Write()
	if str == "" {
		t.Error("expected a output but got an empty string")
		t.FailNow()
	}
	exp := "Age Name\n23  Bob\n50  Alice\n"
	if str != exp {
		t.Error("expect the format to match\n", `[`+exp+`]`, "but got\n", `[`+str+`]`, "instead")
		t.Error("Length of expected", len(exp), "length of string", len(str))
	}

	if len(w.Headers) != 2 {
		t.Error("expect to have 2 headers but got", len(w.Headers), "instead")
		return
	}

	if w.Headers[0] != "Age" {
		t.Error("expect the first header to have the value of [Age] but got", w.Headers[0], "instead")
	}

	if w.Headers[1] != "Name" {
		t.Error("expect the first header to have the value of [Name] but got", w.Headers[0], "instead")
	}
}

func TestComplexWriteToString(t *testing.T) {
	w := wsv.NewWriter()

	r := [][]wsv.RecordField{
		{
			{IsNull: false, Value: "Age", IsHeader: true, FieldIndex: 0, FieldName: "Age", RowIndex: 1},
			{IsNull: false, Value: `Preferred "Nickname"`, IsHeader: true, FieldIndex: 1, FieldName: "Name", RowIndex: 1},
		},
		{
			{IsNull: false, Value: "23", FieldName: "Age", FieldIndex: 0, RowIndex: 2},
			{IsNull: false, Value: `Bob "The Beast" Rogers`, FieldName: `Preferred "Nickname"`, FieldIndex: 1, RowIndex: 2},
		},
		{
			{IsNull: true, Value: "", FieldName: "Age", FieldIndex: 0, RowIndex: 3},
			{IsNull: false, Value: "Alice", FieldName: `Preferred "Nickname"`, FieldIndex: 1, RowIndex: 3},
		},
	}

	for i := range r {
		for _, f := range r[i] {
			err := w.Append(i, f)
			if err != nil {
				t.Error(err)
			}
		}
	}

	if len(w.Records) != 3 {
		t.Error("expected to have 3 rows but got", len(w.Records), "instead")
		t.FailNow()
	}

	str := w.Write()
	if str == "" {
		t.Error("expected a output but got an empty string")
		t.FailNow()
	}
	exp := "Age \"Preferred \"\"Nickname\"\"\"\n23  \"Bob \"\"The Beast\"\" Rogers\"\n-   Alice\n"
	if str != exp {
		t.Error("expect the format to match\n", `[`+exp+`]`, "but got\n", `[`+str+`]`, "instead")
		t.Error("Length of expected", len(exp), "length of string", len(str))
	}

	if len(w.Headers) != 2 {
		t.Error("expect to have 2 headers but got", len(w.Headers), "instead")
		return
	}

	if w.Headers[0] != "Age" {
		t.Error("expect the first header to have the value of [Age] but got", w.Headers[0], "instead")
	}

	if w.Headers[1] != `Preferred "Nickname"` {
		t.Error("expect the first header to have the value of [\"Preferred \"\"Nickname\"\"\"] but got", w.Headers[0], "instead")
	}
}
