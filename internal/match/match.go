package match

import (
	"strings"

	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// Position returns the element's one-based position and sibling count. Names
// containing "of-type" count only siblings with the same tag name.
func Position(name string, e *markup.Element) (pos, total int) {
	if e == nil {
		return 0, 0
	}

	siblings := e.Siblings()
	ofType := strings.Contains(name, "of-type")
	for _, sibling := range siblings {
		if sibling == nil || (ofType && sibling.Name != e.Name) {
			continue
		}
		total++
		if sibling == e {
			pos = total
		}
	}
	return pos, total
}

// SimpleMatches reports whether a single selector matches e.
func SimpleMatches(s selector.Simple, e *markup.Element) bool {
	if e == nil {
		return false
	}

	switch s.Kind {
	case selector.TypeSel:
		return e.Name == s.Name
	case selector.AnySel:
		return true
	case selector.IDSel:
		value, ok := e.Attr("id")
		return ok && value == s.Name
	case selector.ClassSel:
		for _, className := range e.Classes() {
			if className == s.Name {
				return true
			}
		}
		return false
	case selector.AttrSel:
		return attributeMatches(s, e)
	case selector.PseudoSel:
		return pseudoMatches(s, e)
	default:
		return false
	}
}

func attributeMatches(s selector.Simple, e *markup.Element) bool {
	value, exists := e.Attr(s.Name)
	if s.Op == "" {
		return exists
	}
	if !exists {
		return false
	}

	expected := s.Value
	if s.Flag == "i" {
		value = foldASCII(value)
		expected = foldASCII(expected)
	}

	switch s.Op {
	case "=":
		return value == expected
	case "~=":
		if expected == "" || containsBlank(expected) {
			return false
		}
		for _, token := range strings.FieldsFunc(value, matchBlank) {
			if token == expected {
				return true
			}
		}
		return false
	case "|=":
		return value == expected || strings.HasPrefix(value, expected+"-")
	case "^=":
		return expected != "" && strings.HasPrefix(value, expected)
	case "$=":
		return expected != "" && strings.HasSuffix(value, expected)
	case "*=":
		return expected != "" && strings.Contains(value, expected)
	default:
		return false
	}
}

func pseudoMatches(s selector.Simple, e *markup.Element) bool {
	switch s.Name {
	case "root":
		return e.Parent == nil
	case "empty":
		return len(e.Children) == 0 && e.Texts == 0
	case "first-child", "last-child", "only-child",
		"first-of-type", "last-of-type", "only-of-type",
		"nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
		pos, total := Position(s.Name, e)
		switch s.Name {
		case "first-child", "first-of-type":
			return pos == 1
		case "last-child", "last-of-type":
			return pos != 0 && pos == total
		case "only-child", "only-of-type":
			return total == 1
		case "nth-child", "nth-of-type":
			return selector.NthMatches(s.A, s.B, pos)
		case "nth-last-child", "nth-last-of-type":
			return pos != 0 && selector.NthMatches(s.A, s.B, total-pos+1)
		}
	case "not":
		return !ListMatches(s.Args, e)
	case "is", "where":
		return ListMatches(s.Args, e)
	case "has":
		for _, relative := range s.Args {
			if relativeMatches(relative, e) {
				return true
			}
		}
		return false
	}
	return false
}

func containsBlank(text string) bool {
	for i := 0; i < len(text); i++ {
		if matchBlank(rune(text[i])) {
			return true
		}
	}
	return false
}

func matchBlank(ch rune) bool {
	switch ch {
	case ' ', '\t', '\r', '\n', '\f':
		return true
	default:
		return false
	}
}

// foldASCII lower-cases ASCII letters without changing any other bytes.
func foldASCII(text string) string {
	bytes := []byte(text)
	for i, ch := range bytes {
		if ch >= 'A' && ch <= 'Z' {
			bytes[i] = ch + ('a' - 'A')
		}
	}
	return string(bytes)
}

// compoundMatches requires a non-empty compound and every simple selector.
func compoundMatches(compound selector.Compound, e *markup.Element) bool {
	if len(compound) == 0 || e == nil {
		return false
	}
	for _, simple := range compound {
		if !SimpleMatches(simple, e) {
			return false
		}
	}
	return true
}
