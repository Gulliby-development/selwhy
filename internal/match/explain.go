package match

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// Reason returns the first reason a nonmatching selector fails at e. It returns
// an empty string for a matching selector.
func Reason(c *selector.Complex, e *markup.Element) string {
	if e == nil || !validComplex(c) || ComplexMatches(c, e) {
		return ""
	}

	last := len(c.Compounds) - 1
	for _, simple := range c.Compounds[last] {
		if !SimpleMatches(simple, e) {
			return "fails " + base.Cell(selector.SimpleText(simple)) + " (" + failureDetail(simple, e) + ")"
		}
	}

	// If every part of the final compound matched, the relation immediately
	// before it is the first possible failure.
	if last == 0 {
		return ""
	}
	left := &selector.Complex{
		Lead:      c.Lead,
		Compounds: c.Compounds[:last],
		Combs:     c.Combs[:last-1],
	}
	peer := ""
	switch c.Combs[last-1] {
	case ">":
		peer = "parent"
	case " ":
		peer = "ancestor"
	case "+":
		peer = "previous sibling"
	case "~":
		peer = "earlier sibling"
	default:
		return ""
	}
	return "no " + peer + " matches " + base.Cell(selector.ComplexText(left))
}

func failureDetail(simple selector.Simple, e *markup.Element) string {
	switch simple.Kind {
	case selector.TypeSel:
		return "the tag is " + base.Show(e.Name)
	case selector.IDSel:
		if value, ok := e.Attr("id"); ok {
			return "the id is " + base.Show(value)
		}
		return "there is no id"
	case selector.ClassSel:
		return "the classes are " + base.Cell(strings.Join(e.Classes(), " "))
	case selector.AttrSel:
		if value, ok := e.Attr(simple.Name); ok {
			return "the value is " + base.Show(value)
		}
		return "there is no such attribute"
	case selector.PseudoSel:
		switch simple.Name {
		case "empty":
			return "the element has " + strconv.Itoa(directChildCount(e)) + " children"
		case "root":
			return "the element is not the root"
		case "first-child", "last-child", "only-child",
			"first-of-type", "last-of-type", "only-of-type",
			"nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
			pos, total := Position(simple.Name, e)
			return "the position is " + strconv.Itoa(pos) + " of " + strconv.Itoa(total)
		case "not":
			for _, argument := range simple.Args {
				if ComplexMatches(argument, e) {
					return "the element matches " + base.Cell(selector.ComplexText(argument))
				}
			}
			return "the element matches none of the arguments"
		case "is", "where":
			return "the element matches none of the arguments"
		case "has":
			return "no relative selector matches"
		}
	}
	return "the element does not match"
}

// directChildCount counts element children and direct text nodes. Saturating at
// MaxInt keeps the diagnostic safe even for a caller-built Element with an
// inconsistent Texts count.
func directChildCount(e *markup.Element) int {
	count := 0
	for _, child := range e.Children {
		if child == nil {
			continue
		}
		if count == int(^uint(0)>>1) {
			return count
		}
		count++
	}
	if e.Texts <= 0 {
		return count
	}
	maxInt := int(^uint(0) >> 1)
	if count > maxInt-e.Texts {
		return maxInt
	}
	return count + e.Texts
}
