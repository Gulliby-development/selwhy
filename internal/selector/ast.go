package selector

// Kind identifies the kind of a simple selector.
type Kind int

const (
	TypeSel Kind = iota
	AnySel
	IDSel
	ClassSel
	AttrSel
	PseudoSel
)

// Simple is one type, universal, ID, class, attribute, or pseudo-class selector.
type Simple struct {
	Kind  Kind
	Name  string
	Op    string
	Value string
	Flag  string
	Args  List
	A, B  int
}

// Compound is an ordered sequence of simple selectors.
type Compound []Simple

// Complex is a sequence of compounds joined by combinators.
// Combs[i] joins Compounds[i] and Compounds[i+1]. Lead is the optional
// leading combinator of a relative selector.
type Complex struct {
	Lead      string
	Compounds []Compound
	Combs     []string
}

// List is an ordered list of complex selectors.
type List []*Complex
