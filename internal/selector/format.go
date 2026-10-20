package selector

import "strings"

// SimpleText returns the canonical text for a simple selector.
func SimpleText(simple Simple) string {
	return formatSelector(formatAction{kind: formatSimple, simple: simple})
}

// CompoundText returns the canonical text for a compound selector.
func CompoundText(compound Compound) string {
	return formatSelector(formatAction{kind: formatCompound, compound: compound})
}

// ComplexText returns the canonical text for a complex selector.
func ComplexText(complex *Complex) string {
	return formatSelector(formatAction{kind: formatComplex, complex: complex})
}

// ListText returns the canonical text for a selector list.
func ListText(list List) string {
	return formatSelector(formatAction{kind: formatList, list: list})
}

type formatActionKind uint8

const (
	formatText formatActionKind = iota
	formatSimple
	formatCompound
	formatComplex
	formatList
	formatLeaveComplex
)

type formatAction struct {
	kind     formatActionKind
	text     string
	simple   Simple
	compound Compound
	complex  *Complex
	list     List
}

// formatSelector uses an explicit work stack so caller-built nested ASTs cannot
// exhaust the call stack. Active complex selectors are omitted on cycles.
func formatSelector(first formatAction) string {
	var output strings.Builder
	actions := []formatAction{first}
	active := make(map[*Complex]bool)

	for len(actions) != 0 {
		last := len(actions) - 1
		action := actions[last]
		actions = actions[:last]

		switch action.kind {
		case formatText:
			output.WriteString(action.text)

		case formatSimple:
			simple := action.simple
			switch simple.Kind {
			case TypeSel:
				output.WriteString(simple.Name)
			case AnySel:
				output.WriteByte('*')
			case IDSel:
				output.WriteByte('#')
				output.WriteString(simple.Name)
			case ClassSel:
				output.WriteByte('.')
				output.WriteString(simple.Name)
			case AttrSel:
				output.WriteByte('[')
				output.WriteString(simple.Name)
				if simple.Op != "" {
					output.WriteString(simple.Op)
					output.WriteByte('"')
					writeEscapedValue(&output, simple.Value)
					output.WriteByte('"')
					if simple.Flag != "" {
						output.WriteByte(' ')
						output.WriteString(simple.Flag)
					}
				}
				output.WriteByte(']')
			case PseudoSel:
				output.WriteByte(':')
				output.WriteString(simple.Name)
				switch simple.Name {
				case "not", "is", "where", "has":
					actions = append(actions,
						formatAction{kind: formatText, text: ")"},
						formatAction{kind: formatList, list: simple.Args},
						formatAction{kind: formatText, text: "("},
					)
				case "nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
					output.WriteByte('(')
					output.WriteString(NthText(simple.A, simple.B))
					output.WriteByte(')')
				}
			}

		case formatCompound:
			for i := len(action.compound) - 1; i >= 0; i-- {
				actions = append(actions, formatAction{kind: formatSimple, simple: action.compound[i]})
			}

		case formatComplex:
			complex := action.complex
			if complex == nil || active[complex] {
				continue
			}
			active[complex] = true
			actions = append(actions, formatAction{kind: formatLeaveComplex, complex: complex})
			for i := len(complex.Compounds) - 1; i >= 0; i-- {
				actions = append(actions, formatAction{kind: formatCompound, compound: complex.Compounds[i]})
				if i > 0 {
					combinator := " "
					if i-1 < len(complex.Combs) && complex.Combs[i-1] != " " {
						combinator = " " + complex.Combs[i-1] + " "
					}
					actions = append(actions, formatAction{kind: formatText, text: combinator})
				}
			}
			if complex.Lead != "" {
				actions = append(actions, formatAction{kind: formatText, text: complex.Lead + " "})
			}

		case formatList:
			firstMember := true
			for _, complex := range action.list {
				if complex != nil {
					firstMember = false
				}
			}
			if firstMember {
				continue
			}
			for i := len(action.list) - 1; i >= 0; i-- {
				if action.list[i] == nil {
					continue
				}
				actions = append(actions, formatAction{kind: formatComplex, complex: action.list[i]})
				if i > 0 {
					hasPrevious := false
					for j := i - 1; j >= 0; j-- {
						if action.list[j] != nil {
							hasPrevious = true
							break
						}
					}
					if hasPrevious {
						actions = append(actions, formatAction{kind: formatText, text: ", "})
					}
				}
			}

		case formatLeaveComplex:
			delete(active, action.complex)
		}
	}

	return output.String()
}

func writeEscapedValue(output *strings.Builder, value string) {
	for i := 0; i < len(value); i++ {
		if value[i] == '\\' || value[i] == '"' {
			output.WriteByte('\\')
		}
		output.WriteByte(value[i])
	}
}
