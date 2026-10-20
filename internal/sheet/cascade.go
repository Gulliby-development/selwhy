package sheet

import (
	"sort"

	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/match"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// Winner is the winning declaration for one property on an element.
type Winner struct {
	Prop      string
	Value     string
	Important bool
	Rule      int
	Weight    selector.Weight
}

// Cascade resolves each declared property independently for e.
func Cascade(rules []Rule, e *markup.Element) []Winner {
	type candidate struct {
		winner    Winner
		ruleIndex int
		declIndex int
	}

	winners := make(map[string]candidate)
	for ruleIndex, rule := range rules {
		var weight selector.Weight
		matched := false
		for _, complex := range rule.Selectors {
			if !match.ComplexMatches(complex, e) {
				continue
			}
			candidateWeight := selector.Weigh(complex)
			if !matched || weight.Less(candidateWeight) {
				weight = candidateWeight
				matched = true
			}
		}
		if !matched {
			continue
		}

		for declIndex, decl := range rule.Decls {
			current := candidate{
				winner: Winner{
					Prop:      decl.Prop,
					Value:     decl.Value,
					Important: decl.Important,
					Rule:      ruleIndex + 1,
					Weight:    weight,
				},
				ruleIndex: ruleIndex,
				declIndex: declIndex,
			}
			previous, exists := winners[decl.Prop]
			if !exists || outranks(current, previous) {
				winners[decl.Prop] = current
			}
		}
	}

	properties := make([]string, 0, len(winners))
	for prop := range winners {
		properties = append(properties, prop)
	}
	sort.Strings(properties)

	result := make([]Winner, 0, len(properties))
	for _, prop := range properties {
		result = append(result, winners[prop].winner)
	}
	return result
}

func outranks(candidate, previous struct {
	winner    Winner
	ruleIndex int
	declIndex int
}) bool {
	if candidate.winner.Important != previous.winner.Important {
		return candidate.winner.Important
	}
	if candidate.winner.Weight != previous.winner.Weight {
		return previous.winner.Weight.Less(candidate.winner.Weight)
	}
	if candidate.ruleIndex != previous.ruleIndex {
		return candidate.ruleIndex > previous.ruleIndex
	}
	return candidate.declIndex > previous.declIndex
}
