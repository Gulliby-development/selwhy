package match

import (
	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

type matchState struct {
	index   int
	element *markup.Element
}

// ComplexMatches reports whether c matches e, following selector relations
// from right to left and retaining every eligible prefix candidate.
func ComplexMatches(c *selector.Complex, e *markup.Element) bool {
	if !validComplex(c) || e == nil {
		return false
	}
	return matchesPrefix(c, len(c.Compounds)-1, e)
}

// matchesPrefix walks a complex selector from the candidate compound toward
// the left. The work stack performs full-prefix backtracking without recursive
// calls proportional to selector length.
func matchesPrefix(c *selector.Complex, last int, e *markup.Element) bool {
	pending := []matchState{{index: last, element: e}}
	seen := make(map[matchState]struct{})

	for len(pending) != 0 {
		state := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if state.element == nil {
			continue
		}
		if _, visited := seen[state]; visited {
			continue
		}
		seen[state] = struct{}{}
		if state.index < 0 || state.index >= len(c.Compounds) ||
			!compoundMatches(c.Compounds[state.index], state.element) {
			continue
		}
		if state.index == 0 {
			return true
		}

		combinator := c.Combs[state.index-1]
		for _, predecessor := range previousCandidates(state.element, combinator) {
			pending = append(pending, matchState{index: state.index - 1, element: predecessor})
		}
	}
	return false
}

func previousCandidates(e *markup.Element, combinator string) []*markup.Element {
	switch combinator {
	case ">":
		if e.Parent != nil {
			return []*markup.Element{e.Parent}
		}
	case " ":
		var candidates []*markup.Element
		seen := make(map[*markup.Element]struct{})
		for ancestor := e.Parent; ancestor != nil; ancestor = ancestor.Parent {
			if _, visited := seen[ancestor]; visited {
				break
			}
			seen[ancestor] = struct{}{}
			candidates = append(candidates, ancestor)
		}
		return candidates
	case "+", "~":
		siblings := e.Siblings()
		var candidates []*markup.Element
		seen := make(map[*markup.Element]struct{})
		for index, sibling := range siblings {
			if sibling != e {
				continue
			}
			start := index
			if combinator == "+" {
				start--
				if start >= 0 && siblings[start] != nil {
					return []*markup.Element{siblings[start]}
				}
				continue
			}
			for previous := 0; previous < index; previous++ {
				candidate := siblings[previous]
				if candidate == nil {
					continue
				}
				if _, visited := seen[candidate]; !visited {
					seen[candidate] = struct{}{}
					candidates = append(candidates, candidate)
				}
			}
		}
		return candidates
	}
	return nil
}

// ListMatches reports whether any complex selector in l matches e.
func ListMatches(l selector.List, e *markup.Element) bool {
	if e == nil {
		return false
	}
	for _, complex := range l {
		if ComplexMatches(complex, e) {
			return true
		}
	}
	return false
}

// Select returns the ordered, unique document elements matching at least one
// selector in l.
func Select(l selector.List, d *markup.Doc) []*markup.Element {
	if d == nil {
		return nil
	}
	selected := make([]*markup.Element, 0)
	seen := make(map[*markup.Element]struct{})
	for _, element := range d.Elements {
		if element == nil {
			continue
		}
		if _, alreadySelected := seen[element]; alreadySelected {
			continue
		}
		if ListMatches(l, element) {
			seen[element] = struct{}{}
			selected = append(selected, element)
		}
	}
	return selected
}

func validComplex(c *selector.Complex) bool {
	return c != nil && len(c.Compounds) > 0 && len(c.Combs) >= len(c.Compounds)-1
}

// relativeMatches checks a :has argument while keeping the anchor fixed. It
// first binds the selector's first compound to the Lead relation, then walks
// subsequent combinators forward from that element.
func relativeMatches(c *selector.Complex, anchor *markup.Element) bool {
	if !validComplex(c) || anchor == nil {
		return false
	}

	lead := c.Lead
	if lead == "" {
		lead = " "
	}
	firstCandidates := forwardCandidates(anchor, lead)
	pending := make([]matchState, 0, len(firstCandidates))
	for _, candidate := range firstCandidates {
		if compoundMatches(c.Compounds[0], candidate) {
			pending = append(pending, matchState{index: 0, element: candidate})
		}
	}
	seen := make(map[matchState]struct{})

	for len(pending) != 0 {
		state := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if _, visited := seen[state]; visited {
			continue
		}
		seen[state] = struct{}{}
		if state.index == len(c.Compounds)-1 {
			return true
		}

		combinator := c.Combs[state.index]
		for _, candidate := range forwardCandidates(state.element, combinator) {
			if compoundMatches(c.Compounds[state.index+1], candidate) {
				pending = append(pending, matchState{index: state.index + 1, element: candidate})
			}
		}
	}
	return false
}

func forwardCandidates(e *markup.Element, combinator string) []*markup.Element {
	if e == nil {
		return nil
	}
	switch combinator {
	case ">":
		return nonNilElements(e.Children)
	case " ":
		return descendants(e)
	case "+":
		return adjacentFollowingSibling(e)
	case "~":
		return followingSiblings(e)
	}
	return nil
}

func nonNilElements(elements []*markup.Element) []*markup.Element {
	result := make([]*markup.Element, 0, len(elements))
	for _, element := range elements {
		if element != nil {
			result = append(result, element)
		}
	}
	return result
}

func descendants(e *markup.Element) []*markup.Element {
	var result []*markup.Element
	pending := make([]*markup.Element, 0, len(e.Children))
	for i := len(e.Children) - 1; i >= 0; i-- {
		if e.Children[i] != nil {
			pending = append(pending, e.Children[i])
		}
	}
	seen := make(map[*markup.Element]struct{})
	for len(pending) != 0 {
		last := len(pending) - 1
		element := pending[last]
		pending = pending[:last]
		if _, visited := seen[element]; visited {
			continue
		}
		seen[element] = struct{}{}
		result = append(result, element)
		for i := len(element.Children) - 1; i >= 0; i-- {
			if element.Children[i] != nil {
				pending = append(pending, element.Children[i])
			}
		}
	}
	return result
}

func adjacentFollowingSibling(e *markup.Element) []*markup.Element {
	siblings := e.Siblings()
	for index, sibling := range siblings {
		if sibling == e && index+1 < len(siblings) && siblings[index+1] != nil {
			return []*markup.Element{siblings[index+1]}
		}
	}
	return nil
}

func followingSiblings(e *markup.Element) []*markup.Element {
	siblings := e.Siblings()
	var result []*markup.Element
	seen := make(map[*markup.Element]struct{})
	for index, sibling := range siblings {
		if sibling != e {
			continue
		}
		for next := index + 1; next < len(siblings); next++ {
			candidate := siblings[next]
			if candidate == nil {
				continue
			}
			if _, visited := seen[candidate]; !visited {
				seen[candidate] = struct{}{}
				result = append(result, candidate)
			}
		}
	}
	return result
}
