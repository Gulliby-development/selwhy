package markup

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type sourcePos struct {
	line   int
	column int
}

type scanner struct {
	text   string
	offset int
	line   int
	column int
}

func newScanner(text string) scanner {
	return scanner{text: text, line: 1, column: 1}
}

func (s *scanner) eof() bool {
	return s.offset >= len(s.text)
}

func (s *scanner) pos() sourcePos {
	return sourcePos{line: s.line, column: s.column}
}

func (s *scanner) peek() rune {
	if s.eof() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.text[s.offset:])
	return r
}

func (s *scanner) advance() rune {
	if s.eof() {
		return 0
	}
	r, size := utf8.DecodeRuneInString(s.text[s.offset:])
	s.offset += size
	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	return r
}

func (s *scanner) starts(prefix string) bool {
	return strings.HasPrefix(s.text[s.offset:], prefix)
}

type openElement struct {
	element *Element
	start   sourcePos
}

type parseState struct {
	scan  scanner
	doc   *Doc
	stack []openElement
}

// Parse reads the deliberately small markup language used by selwhy.
func Parse(text string) (*Doc, error) {
	p := parseState{scan: newScanner(text), doc: &Doc{}}
	if err := p.parseDocument(); err != nil {
		return nil, err
	}
	return p.doc, nil
}

func (p *parseState) parseDocument() error {
	for !p.scan.eof() {
		if p.scan.peek() != '<' {
			if err := p.parseText(); err != nil {
				return err
			}
			continue
		}

		start := p.scan.pos()
		if p.rootComplete() {
			switch {
			case p.scan.starts("<!--"):
				if err := p.parseComment(); err != nil {
					return err
				}
			case p.scan.starts("</"):
				if err := p.parseEndTag(start); err != nil {
					return err
				}
			default:
				return p.errorAt("more than one root element", start)
			}
			continue
		}

		switch {
		case p.scan.starts("<!--"):
			if err := p.parseComment(); err != nil {
				return err
			}
		case p.scan.starts("<!") || p.scan.starts("<?"):
			return p.errorAt("unsupported markup", start)
		case p.scan.starts("</"):
			if err := p.parseEndTag(start); err != nil {
				return err
			}
		default:
			if p.scan.offset+1 >= len(p.scan.text) || !isASCIINameStart(p.scan.text[p.scan.offset+1]) {
				return p.errorAt("malformed tag", start)
			}
			if len(p.stack) >= 12 {
				return p.errorAt("nesting is deeper than 12", start)
			}
			if err := p.parseStartTag(start); err != nil {
				return err
			}
		}
	}

	if len(p.stack) > 0 {
		open := p.stack[len(p.stack)-1]
		return p.errorAt(fmt.Sprintf("unclosed element %s starting", open.element.Name), open.start)
	}
	if p.doc.Root == nil {
		return fmt.Errorf("no root element")
	}
	return nil
}

func (p *parseState) rootComplete() bool {
	return p.doc.Root != nil && len(p.stack) == 0
}

func (p *parseState) parseText() error {
	inElement := len(p.stack) > 0
	startOffset := p.scan.offset
	runStart := p.scan.pos()
	hasOutsideText := false

	for !p.scan.eof() && p.scan.peek() != '<' {
		r := p.scan.peek()
		if !inElement {
			if !isBlank(r) {
				hasOutsideText = true
			}
			p.scan.advance()
			continue
		}

		if r == '&' {
			if _, err := p.readEntity(); err != nil {
				return err
			}
		} else {
			p.scan.advance()
		}
	}

	if !inElement {
		if hasOutsideText {
			return p.errorAt("text outside the root element", runStart)
		}
		return nil
	}
	if p.scan.offset > startOffset {
		p.stack[len(p.stack)-1].element.Texts++
	}
	return nil
}

func (p *parseState) parseComment() error {
	start := p.scan.pos()
	for i := 0; i < len("<!--"); i++ {
		p.scan.advance()
	}
	end := strings.Index(p.scan.text[p.scan.offset:], "-->")
	if end < 0 {
		return p.errorAt("unterminated comment starting", start)
	}
	end += p.scan.offset + len("-->")
	for p.scan.offset < end {
		p.scan.advance()
	}
	return nil
}

func (p *parseState) parseStartTag(start sourcePos) error {
	p.scan.advance() // <
	name := p.readName()
	element := &Element{Name: name}
	attrs := make([]Attr, 0)
	seen := make(map[string]bool)

	for {
		if p.scan.eof() {
			return p.errorAt("unterminated tag starting", start)
		}
		switch p.scan.peek() {
		case '>':
			p.scan.advance()
			element.Attrs = attrs
			p.appendElement(element, start, false)
			return nil
		case '/':
			if p.scan.offset+1 >= len(p.scan.text) || p.scan.text[p.scan.offset+1] != '>' {
				return p.errorAt("malformed tag", start)
			}
			p.scan.advance()
			p.scan.advance()
			element.Attrs = attrs
			p.appendElement(element, start, true)
			return nil
		default:
			if !isBlank(p.scan.peek()) {
				return p.errorAt("malformed tag", start)
			}
			for !p.scan.eof() && isBlank(p.scan.peek()) {
				p.scan.advance()
			}
			if p.scan.eof() {
				return p.errorAt("unterminated tag starting", start)
			}
			switch p.scan.peek() {
			case '>':
				p.scan.advance()
				element.Attrs = attrs
				p.appendElement(element, start, false)
				return nil
			case '/':
				if p.scan.offset+1 >= len(p.scan.text) || p.scan.text[p.scan.offset+1] != '>' {
					return p.errorAt("malformed tag", start)
				}
				p.scan.advance()
				p.scan.advance()
				element.Attrs = attrs
				p.appendElement(element, start, true)
				return nil
			}

			attr, err := p.parseAttribute(start, seen)
			if err != nil {
				return err
			}
			attrs = append(attrs, attr)
		}
	}
}

func (p *parseState) parseAttribute(tagStart sourcePos, seen map[string]bool) (Attr, error) {
	if !isNameStart(p.scan.peek()) {
		return Attr{}, p.errorAt("malformed tag", tagStart)
	}
	namePos := p.scan.pos()
	name := p.readName()
	if seen[name] {
		return Attr{}, p.errorAt(fmt.Sprintf("duplicate attribute %s", name), namePos)
	}
	seen[name] = true

	afterName := p.scan.offset
	for afterName < len(p.scan.text) && isASCIIBlank(p.scan.text[afterName]) {
		afterName++
	}
	if afterName >= len(p.scan.text) || p.scan.text[afterName] != '=' {
		return Attr{Name: name}, nil
	}

	for p.scan.offset < afterName {
		p.scan.advance()
	}
	p.scan.advance() // =
	for !p.scan.eof() && isBlank(p.scan.peek()) {
		p.scan.advance()
	}
	if p.scan.eof() {
		return Attr{}, p.errorAt("unterminated tag starting", tagStart)
	}

	switch p.scan.peek() {
	case '"', '\'':
		value, err := p.parseQuotedValue()
		if err != nil {
			return Attr{}, err
		}
		return Attr{Name: name, Value: value}, nil
	default:
		value, err := p.parseUnquotedValue(tagStart)
		if err != nil {
			return Attr{}, err
		}
		return Attr{Name: name, Value: value}, nil
	}
}

func (p *parseState) parseQuotedValue() (string, error) {
	quote := p.scan.peek()
	quoteStart := p.scan.pos()
	p.scan.advance()
	var value strings.Builder

	for !p.scan.eof() {
		if p.scan.peek() == quote {
			p.scan.advance()
			return value.String(), nil
		}
		if p.scan.peek() == '&' {
			entity, err := p.readEntity()
			if err != nil {
				return "", err
			}
			value.WriteRune(entity)
			continue
		}
		value.WriteRune(p.scan.advance())
	}
	return "", p.errorAt("unterminated attribute value starting", quoteStart)
}

func (p *parseState) parseUnquotedValue(tagStart sourcePos) (string, error) {
	startOffset := p.scan.offset
	var value strings.Builder
	for !p.scan.eof() {
		r := p.scan.peek()
		if isBlank(r) || r == '"' || r == '\'' || r == '=' || r == '<' || r == '>' || r == '`' {
			break
		}
		if r == '&' {
			entity, err := p.readEntity()
			if err != nil {
				return "", err
			}
			value.WriteRune(entity)
			continue
		}
		value.WriteRune(p.scan.advance())
	}
	if p.scan.offset == startOffset {
		return "", p.errorAt("malformed tag", tagStart)
	}
	return value.String(), nil
}

func (p *parseState) parseEndTag(start sourcePos) error {
	p.scan.advance() // <
	p.scan.advance() // /
	if !isNameStart(p.scan.peek()) {
		return p.errorAt("malformed tag", start)
	}
	name := p.readName()
	for !p.scan.eof() && isBlank(p.scan.peek()) {
		p.scan.advance()
	}
	if p.scan.eof() {
		return p.errorAt("unterminated tag starting", start)
	}
	if p.scan.peek() != '>' {
		return p.errorAt("malformed tag", start)
	}
	p.scan.advance()

	if len(p.stack) == 0 {
		return p.errorAt("unexpected closing tag", start)
	}
	open := p.stack[len(p.stack)-1]
	if name != open.element.Name {
		return p.errorAt(fmt.Sprintf("closing tag </%s> does not match <%s>", name, open.element.Name), start)
	}
	p.stack = p.stack[:len(p.stack)-1]
	return nil
}

func (p *parseState) appendElement(element *Element, start sourcePos, selfClosing bool) {
	element.Num = len(p.doc.Elements) + 1
	if len(p.stack) > 0 {
		parent := p.stack[len(p.stack)-1].element
		element.Parent = parent
		parent.Children = append(parent.Children, element)
	} else {
		p.doc.Root = element
	}
	p.doc.Elements = append(p.doc.Elements, element)
	if !selfClosing {
		p.stack = append(p.stack, openElement{element: element, start: start})
	}
}

func (p *parseState) readName() string {
	var name strings.Builder
	for !p.scan.eof() && isNamePart(p.scan.peek()) {
		name.WriteRune(p.scan.advance())
	}
	return name.String()
}

func (p *parseState) readEntity() (rune, error) {
	start := p.scan.pos()
	semicolon := strings.IndexByte(p.scan.text[p.scan.offset:], ';')
	if semicolon < 0 {
		return 0, p.errorAt("unknown entity", start)
	}
	end := p.scan.offset + semicolon + 1
	raw := p.scan.text[p.scan.offset:end]
	entity, ok := decodeEntity(raw)
	if !ok {
		return 0, p.errorAt("unknown entity", start)
	}
	for p.scan.offset < end {
		p.scan.advance()
	}
	return entity, nil
}

func decodeEntity(raw string) (rune, bool) {
	switch raw {
	case "&amp;":
		return '&', true
	case "&lt;":
		return '<', true
	case "&gt;":
		return '>', true
	case "&quot;":
		return '"', true
	case "&apos;":
		return '\'', true
	}
	if len(raw) < 4 || raw[0] != '&' || raw[1] != '#' || raw[len(raw)-1] != ';' {
		return 0, false
	}

	body := raw[2 : len(raw)-1]
	base := 10
	digits := body
	if strings.HasPrefix(body, "x") {
		base = 16
		digits = body[1:]
		if len(digits) < 1 || len(digits) > 6 {
			return 0, false
		}
	} else if len(digits) < 1 || len(digits) > 7 {
		return 0, false
	}

	value := 0
	for i := 0; i < len(digits); i++ {
		digit := -1
		switch c := digits[i]; {
		case c >= '0' && c <= '9':
			digit = int(c - '0')
		case base == 16 && c >= 'a' && c <= 'f':
			digit = int(c-'a') + 10
		case base == 16 && c >= 'A' && c <= 'F':
			digit = int(c-'A') + 10
		}
		if digit < 0 || digit >= base {
			return 0, false
		}
		value = value*base + digit
	}
	if value < 1 || value > 0x10ffff || (value >= 0xd800 && value <= 0xdfff) {
		return 0, false
	}
	return rune(value), true
}

func (p *parseState) errorAt(message string, pos sourcePos) error {
	return fmt.Errorf("%s at line %d column %d", message, pos.line, pos.column)
}

func isNameStart(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isNamePart(r rune) bool {
	return isNameStart(r) || (r >= '0' && r <= '9') || r == '-'
}

func isASCIINameStart(c byte) bool {
	return c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isASCIIBlank(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '\f':
		return true
	default:
		return false
	}
}
