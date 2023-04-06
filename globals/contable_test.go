package globals

import (
	"fmt"
	"io"
	"os"
	"testing"
)

// ConsolePrintTestStruct is a type that implements IConsolePrint and will be used for testing
type ConsolePrintTestStruct struct {
	Name   string
	Age    string
	Gender string
}

func (td ConsolePrintTestStruct) ColumnCount() int {
	return 4 // the 3 data fields + count
}

func (td ConsolePrintTestStruct) Cell(row int, col int) string {
	switch col {
	case 0:
		return fmt.Sprint(row)
	case 1:
		return td.Name
	case 2:
		return td.Age
	case 3:
		return td.Gender
	}
	panic("bad col passed in")
}

func (td ConsolePrintTestStruct) CellColor(row int, col int) string {
	return ColorReset
}

func (td ConsolePrintTestStruct) FillChar(row int, col int) string {
	return "."
}

/*
this function works by capturing stdout and then printing a table.  it compares to an expected value and errors if they
are different.
*/
func TestPrintTable(t *testing.T) {
	header := []string{"Number", "Name", "Age", "Gender"}
	dataRows := []ConsolePrintTestStruct{
		{"John Doe", "30", "Male"},
		{"Jane Doe", "25", "Female"},
	}

	var rows []IConsolePrint

	for _, r := range dataRows {
		rows = append(rows, r)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := PrintTable(header, rows); err != nil {
		t.Errorf("Error in PrintTable: %v", err)
	}

	w.Close()
	data := make([]byte, 512)
	n, err := r.Read(data)
	os.Stdout = old
	if err != nil && err != io.EOF {
		t.Error("Error reading from pipe:", err)
		return
	}
	output := string(data[:n])
	os.Stdout = old
	expected := "Number  Name      Age  Gender  \n======  ========  ===  ======  \n\x1b[0m0.....  \x1b[0m\x1b[0mJohn Doe  \x1b[0m\x1b[0m30.  \x1b[0m\x1b[0mMale..  \x1b[0m\n\x1b[0m1.....  \x1b[0m\x1b[0mJane Doe  \x1b[0m\x1b[0m25.  \x1b[0m\x1b[0mFemale  \x1b[0m\n"
     fmt.Println(expected)
	if output != expected {
		t.Errorf("Expected output:\n%s\nActual output:\n%s\n", expected, output)
	}
}
