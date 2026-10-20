package markup

import "strings"

// Attr is a source-order name/value pair on an element.
type Attr struct {
	Name  string
	Value string
}

// Element represents one element in a parsed document.
type Element struct {
	Name     string
	Attrs    []Attr
	Parent   *Element
	Children []*Element
	Texts    int
	Num      int
}

// Doc contains the root and all elements in document order.
type Doc struct {
	Root     *Element
	Elements []*Element
}

// Attr returns the value of the first attribute with the given case-sensitive name.
func (e *Element) Attr(name string) (string, bool) {
	if e == nil {
		return "", false
	}
	for _, attr := range e.Attrs {
		if attr.Name == name {
			return attr.Value, true
		}
	}
	return "", false
}

// Classes returns the class attribute split on the five markup blanks.
func (e *Element) Classes() []string {
	if e == nil {
		return nil
	}
	value, ok := e.Attr("class")
	if !ok || value == "" {
		return nil
	}
	return strings.FieldsFunc(value, isBlank)
}

// Siblings returns the parent's element children, or the root itself.
func (e *Element) Siblings() []*Element {
	if e == nil {
		return nil
	}
	if e.Parent == nil {
		return []*Element{e}
	}
	return e.Parent.Children
}

// isBlank reports whether r is one of the blanks accepted by the markup grammar.
func isBlank(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n', '\f':
		return true
	default:
		return false
	}
}
