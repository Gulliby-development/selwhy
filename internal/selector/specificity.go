package selector

import "strconv"

// Weight is the lexicographic specificity tuple (IDs, classes, types).
type Weight [3]int

// Weigh returns the specificity of a complex selector.
func Weigh(c *Complex) Weight {
	if c == nil {
		return Weight{}
	}

	weights := make(map[*Complex]Weight)
	state := make(map[*Complex]uint8) // 1 means visiting; 2 means complete.
	state[c] = 1
	stack := []specificityFrame{{selector: c}}

	for len(stack) > 0 {
		frame := &stack[len(stack)-1]
		if !frame.ready {
			frame.base, frame.groups = specificityParts(frame.selector)
			frame.ready = true
		}

		// Visit nested selector arguments iteratively so caller-built cycles and
		// unusually deep ASTs cannot recurse forever or exhaust the call stack.
		pushed := false
		for frame.group < len(frame.groups) {
			args := frame.groups[frame.group]
			if frame.argument >= len(args) {
				frame.group++
				frame.argument = 0
				continue
			}
			child := args[frame.argument]
			frame.argument++
			if child == nil || state[child] != 0 {
				continue
			}
			state[child] = 1
			stack = append(stack, specificityFrame{selector: child})
			pushed = true
			break
		}
		if pushed {
			continue
		}
		if frame.group < len(frame.groups) {
			continue
		}

		result := frame.base
		for _, args := range frame.groups {
			var greatest Weight
			for _, child := range args {
				if child == nil || state[child] != 2 {
					continue
				}
				candidate := weights[child]
				if greatest.Less(candidate) {
					greatest = candidate
				}
			}
			result = addWeight(result, greatest)
		}

		node := frame.selector
		weights[node] = result
		state[node] = 2
		stack = stack[:len(stack)-1]
	}
	return weights[c]
}

type specificityFrame struct {
	selector *Complex
	base     Weight
	groups   []List
	group    int
	argument int
	ready    bool
}

// specificityParts returns direct contributions and argument lists whose
// maximum specificity contributes in place of the pseudo-class itself.
func specificityParts(c *Complex) (Weight, []List) {
	var weight Weight
	var groups []List
	for _, compound := range c.Compounds {
		for _, simple := range compound {
			switch simple.Kind {
			case TypeSel:
				weight[2] = addSpecificity(weight[2], 1)
			case IDSel:
				weight[0] = addSpecificity(weight[0], 1)
			case ClassSel, AttrSel:
				weight[1] = addSpecificity(weight[1], 1)
			case PseudoSel:
				switch simple.Name {
				case "where":
					// :where() has no specificity.
				case "not", "is", "has":
					groups = append(groups, simple.Args)
				default:
					weight[1] = addSpecificity(weight[1], 1)
				}
			}
		}
	}
	return weight, groups
}

func addWeight(left, right Weight) Weight {
	for i := range left {
		left[i] = addSpecificity(left[i], right[i])
	}
	return left
}

func addSpecificity(left, right int) int {
	maxInt := int(^uint(0) >> 1)
	if right > 0 && left > maxInt-right {
		return maxInt
	}
	return left + right
}

// Less reports whether w has lower specificity than o.
func (w Weight) Less(o Weight) bool {
	for i := range w {
		if w[i] != o[i] {
			return w[i] < o[i]
		}
	}
	return false
}

// String returns the comma-separated decimal coordinates of w.
func (w Weight) String() string {
	return strconv.Itoa(w[0]) + "," + strconv.Itoa(w[1]) + "," + strconv.Itoa(w[2])
}
