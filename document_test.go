package wsv_test

import (
	"io"
	"testing"

	"github.com/internetcalifornia/wsv"
)

func TestCreateDocument(t *testing.T) {
	doc := wsv.NewDocument()

	line, err := doc.AddLine()
	if err != nil {
		t.Error(err)
		return
	}
	if doc.LineCount() != 1 {
		t.Error("Record not updated")
		return
	}
	line.Append("Name")
	line.Append("Age")
	line.Append("Favorite Color")
	line.Append("Preferred \"Nickname\" Name")
	if line.FieldCount() != 4 {
		t.Error("expect line", line.Line(), "to have 4 field but got", line.FieldCount(), "instead")
	}
	field, err := line.NextField()
	if err != nil {
		t.Error(err)
		return
	}
	if field.Value != "Name" || !field.IsHeader || field.FieldName != "Name" {
		t.Errorf("expected the value [Name] but got %+v instead", field)
	}
	field, err = line.NextField()
	if err != nil {
		t.Error(err)
		return
	}
	if field.Value != "Age" || !field.IsHeader || field.FieldName != "Age" {
		t.Errorf("expected the value [Age] but got %+v instead", field)
	}

	doc.AddLine()
	line, err = doc.AddLine()
	if err != nil {
		t.Error(err)
		return
	}

	err = line.Append("Scott")
	if err != nil {
		t.Error(err)
	}
	err = line.Append("33")
	if err != nil {
		t.Error(err)
	}
	err = line.Append("")
	if err != nil {
		t.Error(err)
	}
	err = line.Append("")
	if err != nil {
		t.Error(err)
	}

	err = line.Append("invalid")
	if err == nil {
		t.Error("expected an error since doc is tabular")
	}
	header, err := doc.Line(1)
	if err != nil {
		t.Error(err)
		return
	}
	err = header.UpdateField(0, "Formal Name")
	if err != nil {
		t.Error(err)
		return
	}
	o, err := doc.Write()
	if err != nil {
		t.Error(err)
		return
	}
	exp1 := "\"Formal Name\" Age \"Favorite Color\" \"Preferred \"\"Nickname\"\" Name\"\n"
	if string(o) != exp1 {
		t.Error("expected output to be", []byte(exp1), "but got", o, "instead")
		return
	}
	o, err = doc.Write()
	if err != nil {
		t.Error(err)
		return
	}
	exp2 := "\n"
	if string(o) != exp2 {
		t.Error("expected output to be empty line")
	}
	o, err = doc.Write()
	if err != nil {
		t.Error(err)
		return
	}
	exp3 := "Scott         33  \"\"               \"\"\n"
	if string(o) != exp3 {
		t.Error("expected output to be", []byte(exp3), "but got", o, "instead")
		return
	}

	_, err = doc.Write()
	if err != io.EOF || err == nil {
		t.Error("expected EOF")
	}

	doc.ResetWrite()

	o, err = doc.WriteAll()
	if err != nil {
		t.Error(err)
	}
	exp4 := exp1 + exp2 + exp3
	if string(o) != exp4 {
		t.Error("expected output to be", []byte(exp4), "but got", o, "instead")
	}
}
