package gameofcolor

// Neighbour defines an offset, to describe a neighbouring pixel
type Neighbour struct {
	// X and Y are the offsets relative to the current point
	X, Y int
}

// Rule is a rule for the game
//
// The rule will be constructed as:
//
//	if (Neighbours.haveColor(NeighbourColor) == amount) {
//		thisPixel.Color = NewColor
//	}
type Rule struct {
	// Amount is the amount for which the rule is activated
	Amount int
	// NeighbourColor notes which index of the palette is to be looked at
	NeighbourColor int
	// NewColor is the color this pixel will get if the rule applies
	NewColor int
	// Neightbours points out which neighbours are involved
	Neighbours []Neighbour
}

// RuleMap is a map from int to a slice of Rules
// meaning ColorIndex to the Rules that apply to that colorindex
type RuleMap map[int][]Rule

// EightNeighbours returns the neighbours from GameOfLife
func EightNeighbours() []Neighbour {
	neighbours := make([]Neighbour, 8)

	neighbours[0] = Neighbour{
		X: -1,
		Y: -1,
	}
	neighbours[1] = Neighbour{
		X: -1,
		Y: 0,
	}
	neighbours[2] = Neighbour{
		X: -1,
		Y: 1,
	}
	neighbours[3] = Neighbour{
		X: 0,
		Y: -1,
	}
	neighbours[4] = Neighbour{
		X: 0,
		Y: 1,
	}
	neighbours[5] = Neighbour{
		X: 1,
		Y: -1,
	}
	neighbours[6] = Neighbour{
		X: 1,
		Y: 0,
	}
	neighbours[7] = Neighbour{
		X: 1,
		Y: 1,
	}

	return neighbours
}

// AmountRangeRule adds a range of rules to the RuleMap
func (rm *RuleMap) AmountRangeRule(currColor int, bounds [2]int, nbColor int, nwColor int, nbs []Neighbour) {
	rules := make([]Rule, 0)

	for amount := bounds[0]; amount <= bounds[1]; amount++ {
		rules = append(rules, Rule{
			Amount:         amount,
			NeighbourColor: nbColor,
			NewColor:       nwColor,
			Neighbours:     nbs,
		})
	}

	(*rm)[currColor] = append((*rm)[currColor], rules...)
}

// GameOfLife returns the RuleMap for Conway's Game Of Life
func GameOfLife() RuleMap {
	rulemap := make(RuleMap)

	// rules that apply to black
	rulemap[0] = make([]Rule, 0)

	// if alive, and not 2 or 3 alive neighbours, change to dead
	rulemap.AmountRangeRule(0, [2]int{0, 1}, 0, 1, EightNeighbours())
	rulemap.AmountRangeRule(0, [2]int{4, 8}, 0, 1, EightNeighbours())

	// if dead, spring alive if 3 neighbours
	rulemap[1] = []Rule{
		{
			Amount:         3,
			NeighbourColor: 0,
			NewColor:       0,
			Neighbours:     EightNeighbours(),
		},
	}

	return rulemap
}
