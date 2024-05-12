package document

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestCreateTabularDocument(t *testing.T) {
	doc := NewDocument()

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
	fmt.Printf("%+v\n", line)
	_, err = line.Validate()
	if err != nil {
		t.Error(err)
		return
	}
	if line.FieldCount() != 4 {
		t.Error("expect line", line.LineNumber(), "to have 4 field but got", line.FieldCount(), "instead")
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
	line.UpdateComment("cool person")

	_, err = line.Validate()
	if err != nil {
		t.Error(err)
		return
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
	line, err = doc.AddLine()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("John")
	if err != nil {
		t.Error(err)
		return
	}
	err = line.AppendNull()
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Blue\nGray")
	if err != nil {
		t.Error(err)
		return
	}
	err = line.Append("Johnny\nBoy")
	if err != nil {
		t.Error(err)
		return
	}
	o, err := doc.Write()
	if err != nil {
		t.Error(err)
		return
	}
	exp1 := "\"Formal Name\"  Age  \"Favorite Color\"  \"Preferred \"\"Nickname\"\" Name\"\n"
	if string(o) != exp1 {
		t.Errorf("expected output to be \n%s\nbut got \n%s\ninstead", exp1, string(o))
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
	exp3 := "Scott          33   \"\"                \"\"  #cool person\n"
	if string(o) != exp3 {
		t.Errorf("expected output to be \n%s\nbut got \n%s\ninstead", exp3, string(o))
		return
	}
	exp4 := `John           -    "Blue"/"Gray"     "Johnny"/"Boy"` + string('\n')
	o, err = doc.Write()
	if err != nil {
		t.Error(err)
		return
	}
	if string(o) != exp4 {
		t.Errorf("expected output to be \n%s\nbut got \n%s\ninstead", exp4, string(o))
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
	exp5 := exp1 + exp2 + exp3 + exp4
	if string(o) != exp5 {
		t.Error("expected output to be", []byte(exp5), "but got", o, "instead")
	}
}

func TestStartTabularDocumentWithComment(t *testing.T) {
	doc := NewDocument()
	ln, _ := doc.AddLine()
	ln.UpdateComment("this is a test document")
	_, err := doc.AppendLine(Field("name"), Field("hire date"), Field("salary"))
	if err != nil {
		t.Errorf("failed to write the header line after a document due to %s", err)
		return
	}
	_, err = doc.AppendLine(Field("scott"), Field(time.DateOnly), Field("$200,000,000"))
	if err != nil {
		t.Errorf("failed to write the header line after a document due to %s", err)
		return
	}
	b, err := doc.WriteAll()
	if err != nil {
		t.Errorf("failed to write the bytes due to %s", err)
		return
	}
	if len(b) <= 0 {
		t.Error("expected to have greater than 0 bytes")
	}
}

func TestStartTabularDocumentBeginningWithEmptyLines(t *testing.T) {
	doc := NewDocument()
	doc.AddLine()
	doc.AddLine()
	ln, _ := doc.AddLine()
	ln.UpdateComment("this is a test document")
	_, err := doc.AppendLine(Field("name"), Field("hire date"), Field("salary"))
	if err != nil {
		t.Errorf("failed to write the header line after a document due to %s", err)
		return
	}
	_, err = doc.AppendLine(Field("scott"), Field(time.DateOnly), Field("$200,000,000"))
	if err != nil {
		t.Errorf("failed to write the header line after a document due to %s", err)
		return
	}
	b, err := doc.WriteAll()
	if err != nil {
		t.Errorf("failed to write the bytes due to %s", err)
		return
	}
	if len(b) <= 0 {
		t.Error("expected to have greater than 0 bytes")
	}
}
