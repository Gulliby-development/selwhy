package selector

import (
	"strconv"
	"unicode/utf8"

	"github.com/Gulliby-development/selwhy/internal/base"
)

// Parse parses the documented selector subset, preserving its source order.
func Parse(text string) (List, error) {
	p := selectorParser{text: text, column: 1}
	return p.parseSelectorList(false, false, 0, false)
}

type selectorParser struct {
	text      string
	offset    int
	column    int
	listDepth int
	insideHas bool
}

func (p *selectorParser) eof() bool {
	return p.offset >= len(p.text)
}

func (p *selectorParser) peek() rune {
	if p.eof() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.text[p.offset:])
	return r
}

func (p *selectorParser) advance() rune {
	if p.eof() {
		return 0
	}
	r, size := utf8.DecodeRuneInString(p.text[p.offset:])
	if size == 0 {
		size = 1
	}
	p.offset += size
	p.column++
	return r
}

func (p *selectorParser) skipBlanks() bool {
	start := p.offset
	for !p.eof() && nthBlank(p.text[p.offset]) {
		p.advance()
	}
	return p.offset != start
}

func (p *selectorParser) parseSelectorList(nested, relative bool, depth int, insideHas bool) (List, error) {
	oldDepth, oldHas := p.listDepth, p.insideHas
	p.listDepth, p.insideHas = depth, insideHas
	defer func() {
		p.listDepth, p.insideHas = oldDepth, oldHas
	}()
	var list List
	for {
		p.skipBlanks()
		complex, err := p.parseComplex(relative, nested)
		if err != nil {
			return nil, err
		}
		list = append(list, complex)

		p.skipBlanks()
		if !p.eof() && p.peek() == ',' {
			p.advance()
			continue
		}
		if nested {
			if !p.eof() && p.peek() == ')' {
				p.advance()
				return list, nil
			}
			if p.eof() {
				return nil, selectorError("expected )", p.column)
			}
		} else if p.eof() {
			return list, nil
		}
		if !p.eof() {
			return nil, p.unexpected()
		}
	}
}

func (p *selectorParser) parseComplex(relative, nested bool) (*Complex, error) {
	complex := &Complex{}
	p.skipBlanks()
	if relative && p.isCombinator(p.peek()) {
		complex.Lead = string(p.advance())
		p.skipBlanks()
	}

	compound, err := p.parseCompound()
	if err != nil {
		return nil, err
	}
	complex.Compounds = append(complex.Compounds, compound)

	for {
		hadBlanks := p.skipBlanks()
		ch := p.peek()
		if p.isCombinator(ch) {
			comb := string(p.advance())
			p.skipBlanks()
			next, err := p.parseCompound()
			if err != nil {
				return nil, err
			}
			complex.Combs = append(complex.Combs, comb)
			complex.Compounds = append(complex.Compounds, next)
			continue
		}
		if hadBlanks && p.isCompoundStart() {
			next, err := p.parseCompound()
			if err != nil {
				return nil, err
			}
			complex.Combs = append(complex.Combs, " ")
			complex.Compounds = append(complex.Compounds, next)
			continue
		}
		if p.eof() || ch == ',' || (nested && ch == ')') {
			return complex, nil
		}
		return nil, p.unexpected()
	}
}

func (p *selectorParser) parseCompound() (Compound, error) {
	var compound Compound

	if p.peek() == '*' {
		p.advance()
		compound = append(compound, Simple{Kind: AnySel})
	} else if p.identifierStartsAt(p.offset) {
		name, _ := p.parseIdentifier()
		compound = append(compound, Simple{Kind: TypeSel, Name: name})
	}

	for !p.eof() {
		switch p.peek() {
		case '#':
			p.advance()
			name, ok := p.parseIdentifier()
			if !ok {
				return nil, selectorError("expected a name", p.column)
			}
			compound = append(compound, Simple{Kind: IDSel, Name: name})
		case '.':
			p.advance()
			name, ok := p.parseIdentifier()
			if !ok {
				return nil, selectorError("expected a name", p.column)
			}
			compound = append(compound, Simple{Kind: ClassSel, Name: name})
		case '[':
			simple, err := p.parseAttribute()
			if err != nil {
				return nil, err
			}
			compound = append(compound, simple)
		case ':':
			simple, err := p.parsePseudo()
			if err != nil {
				return nil, err
			}
			compound = append(compound, simple)
		default:
			if len(compound) == 0 {
				return nil, selectorError("expected a selector", p.column)
			}
			return compound, nil
		}
	}

	if len(compound) == 0 {
		return nil, selectorError("expected a selector", p.column)
	}
	return compound, nil
}

func (p *selectorParser) parseAttribute() (Simple, error) {
	p.advance() // '['
	p.skipBlanks()
	name, ok := p.parseIdentifier()
	if !ok {
		return Simple{}, selectorError("expected a name", p.column)
	}
	simple := Simple{Kind: AttrSel, Name: name}

	p.skipBlanks()
	if !p.eof() && p.peek() == ']' {
		p.advance()
		return simple, nil
	}
	if p.eof() {
		return Simple{}, selectorError("expected ]", p.column)
	}

	ch := p.peek()
	switch ch {
	case '=':
		simple.Op = "="
		p.advance()
	case '~', '|', '^', '$', '*':
		p.advance()
		if p.eof() || p.peek() != '=' {
			return Simple{}, selectorError("expected =", p.column)
		}
		simple.Op = string(ch) + "="
		p.advance()
	default:
		return Simple{}, p.unexpected()
	}

	p.skipBlanks()
	if p.eof() || p.peek() == ']' {
		return Simple{}, selectorError("expected an attribute value", p.column)
	}
	if p.peek() == '"' || p.peek() == '\'' {
		value, err := p.parseString()
		if err != nil {
			return Simple{}, err
		}
		simple.Value = value
	} else {
		value, ok := p.parseIdentifier()
		if !ok {
			return Simple{}, selectorError("expected an attribute value", p.column)
		}
		simple.Value = value
	}

	p.skipBlanks()
	if !p.eof() && (p.peek() == 'i' || p.peek() == 's') {
		flag := string(p.advance())
		if !p.eof() && !nthBlank(p.text[p.offset]) && p.peek() != ']' {
			return Simple{}, p.unexpected()
		}
		simple.Flag = flag
		p.skipBlanks()
	}
	if !p.eof() && p.peek() == ']' {
		p.advance()
		return simple, nil
	}
	if p.eof() {
		return Simple{}, selectorError("expected ]", p.column)
	}
	return Simple{}, p.unexpected()
}

func (p *selectorParser) parseString() (string, error) {
	quote := p.advance()
	startColumn := p.column - 1
	var value []rune
	for !p.eof() {
		ch := p.peek()
		if ch == '\n' {
			return "", selectorError("unterminated string starting", startColumn)
		}
		if ch == quote {
			p.advance()
			return string(value), nil
		}
		if ch == '\\' {
			slashColumn := p.column
			p.advance()
			if p.eof() {
				return "", selectorError("unterminated string starting", startColumn)
			}
			escaped := p.peek()
			if escaped != '"' && escaped != '\'' && escaped != '\\' {
				return "", selectorError("unsupported escape", slashColumn)
			}
			value = append(value, escaped)
			p.advance()
			continue
		}
		value = append(value, p.advance())
	}
	return "", selectorError("unterminated string starting", startColumn)
}

func (p *selectorParser) parsePseudo() (Simple, error) {
	colonColumn := p.column
	p.advance()
	if !p.eof() && p.peek() == ':' {
		return Simple{}, selectorError("pseudo-elements are not supported", colonColumn)
	}
	name, ok := p.parseIdentifier()
	if !ok {
		return Simple{}, selectorError("expected a name", p.column)
	}

	simple := Simple{Kind: PseudoSel, Name: name}
	if isNoArgumentPseudo(name) {
		if !p.eof() && p.peek() == '(' {
			return Simple{}, selectorError(":"+name+" takes no arguments", colonColumn)
		}
		return simple, nil
	}
	if isFormulaPseudo(name) {
		if p.eof() || p.peek() != '(' {
			return Simple{}, selectorError(":"+name+" needs arguments", colonColumn)
		}
		p.advance()
		a, b, err := p.parseFormula()
		if err != nil {
			return Simple{}, err
		}
		simple.A, simple.B = a, b
		return simple, nil
	}
	if !isListPseudo(name) {
		return Simple{}, selectorError("unknown pseudo-class "+base.Show(":"+name), colonColumn)
	}
	if p.eof() || p.peek() != '(' {
		return Simple{}, selectorError(":"+name+" needs arguments", colonColumn)
	}
	if p.listDepth >= 8 {
		return Simple{}, selectorError("nesting is deeper than 8", colonColumn)
	}
	if name == "has" && p.insideHas {
		return Simple{}, selectorError(":has() cannot be nested in :has()", colonColumn)
	}
	p.advance() // '('
	args, err := p.parseSelectorList(true, name == "has", p.listDepth+1, p.insideHas || name == "has")
	if err != nil {
		return Simple{}, err
	}
	simple.Args = args
	return simple, nil
}

func (p *selectorParser) parseFormula() (int, int, error) {
	startOffset := p.offset
	startColumn := p.column
	for !p.eof() && p.peek() != ')' {
		p.advance()
	}
	if p.eof() {
		return 0, 0, selectorError("expected )", p.column)
	}

	endOffset := p.offset
	closeColumn := p.column
	raw := p.text[startOffset:endOffset]
	left, right := 0, len(raw)
	for left < right && nthBlank(raw[left]) {
		left++
	}
	for right > left && nthBlank(raw[right-1]) {
		right--
	}
	trimmed := raw[left:right]
	column := startColumn + left
	if trimmed == "" {
		column = closeColumn
	}
	a, b, ok := ParseNth(raw)
	p.advance() // ')'
	if !ok {
		return 0, 0, selectorError("invalid an+b formula "+base.Show(trimmed), column)
	}
	return a, b, nil
}

func (p *selectorParser) parseIdentifier() (string, bool) {
	if !p.identifierStartsAt(p.offset) {
		return "", false
	}
	start := p.offset
	if p.peek() == '-' {
		p.advance()
	}
	p.advance() // first ASCII letter or underscore
	for !p.eof() && isIdentifierTail(p.text[p.offset]) {
		p.advance()
	}
	return p.text[start:p.offset], true
}

func (p *selectorParser) identifierStartsAt(offset int) bool {
	if offset >= len(p.text) {
		return false
	}
	if p.text[offset] == '-' {
		offset++
		if offset >= len(p.text) {
			return false
		}
	}
	ch := p.text[offset]
	return isASCIILetter(ch) || ch == '_'
}

func (p *selectorParser) isCompoundStart() bool {
	if p.eof() {
		return false
	}
	switch p.peek() {
	case '*', '#', '.', '[', ':':
		return true
	default:
		return p.identifierStartsAt(p.offset)
	}
}

func (p *selectorParser) isCombinator(ch rune) bool {
	return ch == '>' || ch == '+' || ch == '~'
}

func (p *selectorParser) unexpected() error {
	if p.eof() {
		return selectorError("expected a selector", p.column)
	}
	return selectorError("unexpected "+base.Show(string(p.peek())), p.column)
}

func selectorError(message string, column int) error {
	return selectorErrorColumn(message, column)
}

func selectorErrorColumn(message string, column int) error {
	return &parseError{message: message, column: column}
}

type parseError struct {
	message string
	column  int
}

func (e *parseError) Error() string {
	return e.message + " at column " + strconv.Itoa(e.column)
}

func isNoArgumentPseudo(name string) bool {
	switch name {
	case "first-child", "last-child", "only-child", "first-of-type",
		"last-of-type", "only-of-type", "empty", "root":
		return true
	default:
		return false
	}
}

func isFormulaPseudo(name string) bool {
	switch name {
	case "nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
		return true
	default:
		return false
	}
}

func isListPseudo(name string) bool {
	switch name {
	case "not", "is", "where", "has":
		return true
	default:
		return false
	}
}

func isASCIILetter(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isIdentifierTail(ch byte) bool {
	return isASCIILetter(ch) || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}
