package base

import (
	"strings"
	"unicode/utf8"
)

// Show wraps s in double quotes, replacing each non-printable ASCII rune with ?.
func Show(s string) string {
	return asciiWrapped(s, '"', '"')
}

// Cell wraps s in square brackets, replacing each non-printable ASCII rune with ?.
func Cell(s string) string {
	return asciiWrapped(s, '[', ']')
}

func asciiWrapped(s string, open, close byte) string {
	var b strings.Builder
	b.WriteByte(open)
	for _, r := range s {
		if r < ' ' || r > '~' {
			b.WriteByte('?')
		} else {
			b.WriteByte(byte(r))
		}
	}
	b.WriteByte(close)
	return b.String()
}

// Line formats a key/value pair, padding the key to at least ten code points.
func Line(key, value string) string {
	var b strings.Builder
	b.WriteString(key)
	for n := utf8.RuneCountInString(key); n < 10; n++ {
		b.WriteByte(' ')
	}
	b.WriteString("  ")
	b.WriteString(value)
	return b.String()
}

// Table formats a header and rows into aligned, left-justified lines.
func Table(header []string, rows [][]string) []string {
	columnCount := len(header)
	for _, row := range rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}

	widths := make([]int, columnCount)
	for col, cell := range header {
		widths[col] = utf8.RuneCountInString(cell)
	}
	for _, row := range rows {
		for col, cell := range row {
			width := utf8.RuneCountInString(cell)
			if width > widths[col] {
				widths[col] = width
			}
		}
	}

	lines := make([]string, 0)
	lines = append(lines, tableLine(header, widths))
	for _, row := range rows {
		lines = append(lines, tableLine(row, widths))
	}
	return lines
}

func tableLine(cells []string, widths []int) string {
	var b strings.Builder
	for col, width := range widths {
		if col > 0 {
			b.WriteString("  ")
		}

		cell := ""
		if col < len(cells) {
			cell = cells[col]
		}
		b.WriteString(cell)

		if col < len(widths)-1 {
			for n := utf8.RuneCountInString(cell); n < width; n++ {
				b.WriteByte(' ')
			}
		}
	}
	return strings.TrimRight(b.String(), " ")
}
