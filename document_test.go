package wsv_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/internetcalifornia/wsv"
)

func TestCreateTabularDocument(t *testing.T) {
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
	_, err = line.Validate()
	if err != nil {
		t.Error(err)
		return
	}
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
	line.Comment = "cool person"

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

func TestRuneCounting(t *testing.T) {
	str := `"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."`
	c := 0

	for i, r := range str {
		if i != c {
			t.Error("Out of sync", i, c, r, string(r), utf8.RuneCountInString(str))
		}
		c++
	}
}

func TestSerializeText(t *testing.T) {
	rec1 := wsv.RecordField{
		Value: "Japan is a volcanic archipelago with over 100 active volcanoes.\nThe currency is the yen and the symbol is ¥.",
	}
	exp1 := `"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."`
	out1 := rec1.SerializeText()
	if out1 != exp1 {
		t.Errorf("expect\n%s\nbut got\n%s\ninstead", exp1, out1)
	}
	cal1 := wsv.CalculateFieldLength(rec1)
	if cal1 != 112 {
		t.Error(cal1)
	}

	rec2 := wsv.RecordField{Value: "Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!"}
	exp2 := `"Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!"`
	out2 := rec2.SerializeText()
	if out2 != exp2 {
		t.Errorf("expect\n%s\nbut got\n%s\ninstead", exp2, out2)
	}

}

func TestWriteComplexLine(t *testing.T) {
	strs := []string{}
	strs = append(strs, `Country						Capital    	        "Emoji of Flag" "Interesting Facts" 																								#facts generated from Google's Gemini 2024-04-24`)
	strs = append(strs, `Japan						Tokyo				🇯🇵🇯🇵			"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."   #has half-width characters`)

	r := wsv.NewReader(strings.NewReader(strings.Join(strs, "\n")))

	lines, err := r.ReadAll()
	if err != nil {
		t.Error(err)
		return
	}
	if len(lines) != 2 {
		t.Error("expected 2 lines but got", len(lines))
		return
	}
	doc := wsv.NewDocument()
	doc.Tabular = true
	doc.SetPadding([]rune{' ', ' '})
	for n, line := range lines {

		ln, err := doc.AddLine()
		if err != nil {
			t.Error(err)
			return
		}
		for fi, field := range line.Fields {
			if n == 1 && fi == 3 && field.SerializeText() != `"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."` {
				t.Error(field.SerializeText())
				return
			}
			err := ln.Append(field.Value)
			if err != nil {
				t.Error(err)
				return
			}
		}
		if line.Comment != "" {
			ln.Comment = line.Comment
		}
	}
	if doc.LineCount() != 2 {
		t.Error("expected doc to have 2 lines but got", doc.LineCount())
	}

	data, err := doc.WriteAll()
	if err != nil {
		t.Error(err)
		return
	}
	file, err := os.Create(fmt.Sprintf("%s/example-output/simple-output.wsv", basepath))
	if err != nil {
		t.Error(err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		t.Error(err)
	}
}
