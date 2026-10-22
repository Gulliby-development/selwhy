package markup

import (
	"strconv"
	"strings"
)

// Path returns the slash-separated path from the root to e.
func Path(e *Element) string {
	if e == nil {
		return ""
	}

	chain := make([]*Element, 0)
	for current := e; current != nil; current = current.Parent {
		chain = append(chain, current)
	}

	var path strings.Builder
	for i := len(chain) - 1; i >= 0; i-- {
		element := chain[i]
		position := 1
		siblings := element.Siblings()
		for _, sibling := range siblings {
			if sibling == element {
				break
			}
			if sibling.Name == element.Name {
				position++
			}
		}
		path.WriteByte('/')
		path.WriteString(element.Name)
		path.WriteByte('[')
		path.WriteString(strconv.Itoa(position))
		path.WriteByte(']')
	}
	return path.String()
}

// Depth returns the number of elements from the root to e, both counted.
func Depth(e *Element) int {
	depth := 0
	for current := e; current != nil; current = current.Parent {
		depth++
	}
	return depth
}

// TagText renders e's start tag with attributes in their source order.
func TagText(e *Element) string {
	if e == nil {
		return ""
	}
	var tag strings.Builder
	tag.WriteByte('<')
	tag.WriteString(e.Name)
	for _, attr := range e.Attrs {
		tag.WriteByte(' ')
		tag.WriteString(attr.Name)
		tag.WriteString("=\"")
		tag.WriteString(attr.Value)
		tag.WriteByte('"')
	}
	tag.WriteByte('>')
	return tag.String()
}
