package selector

import "strconv"

// ParseNth parses the bounded an+b syntax used by the selector subset.
func ParseNth(arg string) (a, b int, ok bool) {
	start, end := 0, len(arg)
	for start < end && nthBlank(arg[start]) {
		start++
	}
	for start < end && nthBlank(arg[end-1]) {
		end--
	}
	text := arg[start:end]
	if text == "" {
		return 0, 0, false
	}
	if text == "odd" {
		return 2, 1, true
	}
	if text == "even" {
		return 2, 0, true
	}

	i := 0
	negative := false
	if text[i] == '+' || text[i] == '-' {
		negative = text[i] == '-'
		i++
	}
	digitStart := i
	for i < len(text) && nthDigit(text[i]) {
		i++
	}
	digitCount := i - digitStart

	// A signed integer has at least one digit and no trailing characters.
	if i == len(text) {
		if digitCount == 0 || digitCount > 6 {
			return 0, 0, false
		}
		value := nthDigits(text[digitStart:i])
		if negative {
			value = -value
		}
		return 0, value, true
	}

	// A coefficient is optional, but when present it is limited to six digits.
	if digitCount > 6 || i >= len(text) || text[i] != 'n' {
		return 0, 0, false
	}
	if digitCount == 0 {
		a = 1
	} else {
		a = nthDigits(text[digitStart:i])
	}
	if negative {
		a = -a
	}
	i++

	if i == len(text) {
		return a, 0, true
	}

	for i < len(text) && nthBlank(text[i]) {
		i++
	}
	if i == len(text) || (text[i] != '+' && text[i] != '-') {
		return 0, 0, false
	}
	offsetNegative := text[i] == '-'
	i++
	for i < len(text) && nthBlank(text[i]) {
		i++
	}
	offsetStart := i
	for i < len(text) && nthDigit(text[i]) {
		i++
	}
	offsetCount := i - offsetStart
	if offsetCount == 0 || offsetCount > 6 || i != len(text) {
		return 0, 0, false
	}
	b = nthDigits(text[offsetStart:i])
	if offsetNegative {
		b = -b
	}
	return a, b, true
}

// NthMatches reports whether a nonnegative integer n satisfies a*n+b=pos.
// The calculation uses unsigned differences so the full int range is safe.
func NthMatches(a, b, pos int) bool {
	if a == 0 {
		return pos == b
	}
	if a > 0 {
		if pos < b {
			return false
		}
		difference := uint(pos) - uint(b)
		return difference%uint(a) == 0
	}
	if pos > b {
		return false
	}
	difference := uint(b) - uint(pos)
	magnitude := uint(-(a + 1)) + 1
	return difference%magnitude == 0
}

// NthText returns the canonical textual representation of an+b.
func NthText(a, b int) string {
	var text string
	switch a {
	case 0:
		return strconv.Itoa(b)
	case 1:
		text = "n"
	case -1:
		text = "-n"
	default:
		text = strconv.Itoa(a) + "n"
	}
	if b > 0 {
		return text + "+" + strconv.Itoa(b)
	}
	if b < 0 {
		return text + strconv.Itoa(b)
	}
	return text
}

func nthBlank(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '\f'
}

func nthDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// nthDigits converts a previously validated, at-most-six-digit field.
func nthDigits(text string) int {
	value := 0
	for i := 0; i < len(text); i++ {
		value = value*10 + int(text[i]-'0')
	}
	return value
}
