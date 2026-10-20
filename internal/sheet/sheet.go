package sheet

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// Decl is one property/value pair in source order.
type Decl struct {
	Prop      string
	Value     string
	Important bool
}

// Rule pairs its ordered selector list with its ordered declarations.
type Rule struct {
	Selectors selector.List
	Decls     []Decl
}

// Parse parses a style sheet in the documented subset.
func Parse(text string) ([]Rule, error) {
	normalized, err := removeComments(text)
	if err != nil {
		return nil, err
	}

	var rules []Rule
	offset := 0
	for {
		offset = skipBlanks(normalized, offset)
		if offset >= len(normalized) {
			break
		}

		ruleNumber := len(rules) + 1
		if normalized[offset] == '@' {
			return nil, ruleError(ruleNumber, "at-rules are not supported")
		}

		open := strings.IndexByte(normalized[offset:], '{')
		if open < 0 {
			return nil, ruleError(ruleNumber, "missing {")
		}
		open += offset

		selectorText := trimBlanks(normalized[offset:open])
		selectors, err := selector.Parse(selectorText)
		if err != nil {
			return nil, ruleError(ruleNumber, err.Error())
		}

		closeRelative := strings.IndexByte(normalized[open+1:], '}')
		if closeRelative < 0 {
			return nil, ruleError(ruleNumber, "missing }")
		}
		close := open + 1 + closeRelative
		if strings.IndexByte(normalized[open+1:close], '{') >= 0 {
			return nil, ruleError(ruleNumber, "unexpected {")
		}

		decls, err := parseDeclarations(normalized[open+1:close], ruleNumber)
		if err != nil {
			return nil, err
		}
		rules = append(rules, Rule{Selectors: selectors, Decls: decls})
		offset = close + 1
	}

	if len(rules) == 0 {
		return nil, errors.New("the style sheet has no rules")
	}
	return rules, nil
}

func ruleError(ruleNumber int, message string) error {
	return errors.New("rule " + strconv.Itoa(ruleNumber) + ": " + message)
}

func parseDeclarations(body string, ruleNumber int) ([]Decl, error) {
	var decls []Decl
	declarationNumber := 0
	start := 0

	for start <= len(body) {
		end := start
		for end < len(body) && body[end] != ';' {
			end++
		}
		part := trimBlanks(body[start:end])
		if part != "" {
			declarationNumber++
			colon := strings.IndexByte(part, ':')
			if colon < 0 {
				return nil, ruleError(ruleNumber, "declaration "+strconv.Itoa(declarationNumber)+" has no colon")
			}

			prop := trimBlanks(part[:colon])
			if !validProperty(prop) {
				return nil, ruleError(ruleNumber, "invalid property name "+base.Show(prop))
			}
			value := trimBlanks(part[colon+1:])
			value, important := stripImportant(value)
			if value == "" {
				return nil, ruleError(ruleNumber, "property "+prop+" has no value")
			}
			decls = append(decls, Decl{Prop: prop, Value: value, Important: important})
		}

		if end == len(body) {
			break
		}
		start = end + 1
	}
	return decls, nil
}

func stripImportant(value string) (string, bool) {
	const suffix = "important"
	if !strings.HasSuffix(value, suffix) {
		return value, false
	}
	marker := len(value) - len(suffix) - 1
	if marker < 0 {
		return value, false
	}
	for marker >= 0 && isBlank(value[marker]) {
		marker--
	}
	if marker < 0 || value[marker] != '!' {
		return value, false
	}
	return trimBlanks(value[:marker]), true
}

func validProperty(prop string) bool {
	if prop == "" {
		return false
	}
	for index := 0; index < len(prop); index++ {
		ch := prop[index]
		if index == 0 {
			if !isASCIILetter(ch) && ch != '_' && ch != '-' {
				return false
			}
			continue
		}
		if !isASCIILetter(ch) && (ch < '0' || ch > '9') && ch != '_' && ch != '-' {
			return false
		}
	}
	return true
}

func removeComments(text string) (string, error) {
	var result strings.Builder
	result.Grow(len(text))
	offset := 0
	line, column := 1, 1
	for {
		relative := strings.Index(text[offset:], "/*")
		if relative < 0 {
			result.WriteString(text[offset:])
			return result.String(), nil
		}
		start := offset + relative
		line, column = advancePosition(text[offset:start], line, column)
		closeRelative := strings.Index(text[start+2:], "*/")
		if closeRelative < 0 {
			return "", errors.New("unterminated comment starting at line " +
				strconv.Itoa(line) + " column " + strconv.Itoa(column))
		}
		close := start + 2 + closeRelative
		result.WriteString(text[offset:start])
		result.WriteByte(' ')
		offset = close + 2
		line, column = advancePosition(text[start:offset], line, column)
	}
}

func advancePosition(text string, line, column int) (int, int) {
	for _, r := range text {
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func skipBlanks(text string, offset int) int {
	for offset < len(text) && isBlank(text[offset]) {
		offset++
	}
	return offset
}

func trimBlanks(text string) string {
	start, end := 0, len(text)
	for start < end && isBlank(text[start]) {
		start++
	}
	for end > start && isBlank(text[end-1]) {
		end--
	}
	return text[start:end]
}

func isBlank(ch byte) bool {
	switch ch {
	case ' ', '\t', '\r', '\n', '\f':
		return true
	default:
		return false
	}
}

func isASCIILetter(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}
