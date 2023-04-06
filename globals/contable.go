package globals

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// IConsolePrint defines the interface for printing console tables
type IConsolePrint interface {
	ColumnCount() int                  // ColumnCount returns the number of columns in the table
	Cell(row int, col int) string      // Cell returns the string to print for the specified row and column
	CellColor(row int, col int) string // CellColor returns the color to print the cell
	FillChar(row int, col int) string  // FillChar returns the character to fill the space in a cell
}

// PrintTable prints a table to the console with the given header and rows
func PrintTable(header []string, rows []IConsolePrint) (err error) {
	columnGap := "  "
	columnCount := len(header)
	var widths = make([]int, columnCount)

	// Initialize widths with the width of the header
	for col := 0; col < columnCount; col++ {
		valueLen := utf8.RuneCountInString(header[col])
		if valueLen > widths[col] {
			widths[col] = valueLen
		}
	}

	// Find the longest value for each column by going through each row
	rowCount := len(rows)
	for col := 0; col < columnCount; col++ {
		for row := 0; row < rowCount; row++ {
			valueLength := utf8.RuneCountInString(rows[row].Cell(row, col))
			if valueLength > widths[col] {
				widths[col] = valueLength
			}
		}
	}

	// Print the header
	for col := 0; col < columnCount; col++ {
		fmt.Print(header[col], strings.Repeat(" ", widths[col]-utf8.RuneCountInString(header[col])), columnGap)
	}
	fmt.Print("\n")

	// Print a separator
	for col := 0; col < columnCount; col++ {
		fmt.Print(strings.Repeat("=", widths[col]), columnGap)
	}
	fmt.Print("\n")

	// Print the rows
	for row := 0; row < rowCount; row++ {
		for col := 0; col < columnCount; col++ {
			value := rows[row].Cell(row, col)
			fillChar := rows[row].FillChar(row, col)
			fmt.Print(rows[row].CellColor(row, col), value, strings.Repeat(fillChar, widths[col]-utf8.RuneCountInString(value)), columnGap, ColorReset)
		}
		fmt.Print("\n")
	}

	return
}
