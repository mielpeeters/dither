// Package gameofcolor is a generalization of the Conway's Game of Life idea.
//
// Users can create their own rules, as well as use the rules that are pre-defined,
// to create a GIF image of some rules being applied to an inputted image.
package gameofcolor

// TODO: find a way to combine multiple rules, based on some probability or weight maybe
// TODO: define different types of rules that adjust the neighouring pixels in some way,
//		 this could create movement possibly

import (
	"image"
	"runtime"
	"sync"

	"github.com/mielpeeters/dither/gifeo"
	"github.com/mielpeeters/dither/needle"
	"github.com/mielpeeters/pacebar"
)

// Neighbour defines an offset, to describe a neighbouring pixel
type Neighbour struct {
	// X and Y are the offsets relative to the current point
	X, Y int
}

// Rule is a rule for the game
//
// The rule will be constructed as:
//
//		if Lower <= Neighbours.haveColor(NeighbourColor) <= Upper {
//			thisPixel.Color = NewColorIf
//		} else {
//			thisPixel.Color = NewColorElse
//	}
type Rule struct {
	// Lower and Upper define the count range for which the rule applies
	Lower, Upper int
	// NeighbourColor notes which index of the palette is to be looked at
	NeighbourColor uint8
	// NewColorIf is the color this pixel will get if the rule applies
	NewColorIf uint8
	// NewColorElse is the color this pixel will get if the rule doesn't apply
	NewColorElse uint8
	// Neightbours points out which neighbours are involved
	Neighbours []Neighbour
}

// RuleMap is a map from int to a slice of Rules
// meaning ColorIndex to the Rules that apply to that colorindex
type RuleMap map[uint8][]Rule

// NeighbourLeft returns the left neighbour set
func NeighbourLeft() []Neighbour {
	return []Neighbour{
		{
			X: -1,
			Y: 0,
		},
	}
}

// NeighbourRight returns the right neighbour set
func NeighbourRight() []Neighbour {
	return []Neighbour{
		{
			X: 1,
			Y: 0,
		},
	}
}

// NeighboursLeft returns the left neighbours set
func NeighboursLeft() []Neighbour {
	return []Neighbour{
		{
			X: -1,
			Y: 0,
		},
		{
			X: -2,
			Y: 0,
		},
		{
			X: -1,
			Y: 1,
		},
		{
			X: -1,
			Y: -1,
		},
	}
}

// NeighboursRight returns the right neighbours set
func NeighboursRight() []Neighbour {
	return []Neighbour{
		{
			X: 1,
			Y: 0,
		},
		{
			X: 2,
			Y: 0,
		},
		{
			X: 1,
			Y: 1,
		},
		{
			X: 1,
			Y: -1,
		},
	}
}

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

// TwelveNeighbours returns the neighbours from GameOfLife, with points added
func TwelveNeighbours() []Neighbour {
	neighbours := make([]Neighbour, 12)

	neighbours[0] = Neighbour{X: -1, Y: -1}
	neighbours[1] = Neighbour{X: -1, Y: 0}
	neighbours[2] = Neighbour{X: -1, Y: 1}
	neighbours[3] = Neighbour{X: 0, Y: -1}
	neighbours[4] = Neighbour{X: 0, Y: 1}
	neighbours[5] = Neighbour{X: 1, Y: -1}
	neighbours[6] = Neighbour{X: 1, Y: 0}
	neighbours[7] = Neighbour{X: 1, Y: 1}
	neighbours[8] = Neighbour{X: 0, Y: 2}
	neighbours[9] = Neighbour{X: 2, Y: 0}
	neighbours[10] = Neighbour{X: -2, Y: 0}
	neighbours[11] = Neighbour{X: 0, Y: -2}

	return neighbours
}

// CircleNeighbours returns the neighbours within a close circle
func CircleNeighbours() []Neighbour {
	neighbours := make([]Neighbour, 20)

	neighbours[0] = Neighbour{X: -1, Y: -2}
	neighbours[1] = Neighbour{X: 0, Y: -2}
	neighbours[2] = Neighbour{X: 1, Y: -2}

	neighbours[3] = Neighbour{X: -2, Y: -1}
	neighbours[4] = Neighbour{X: -1, Y: -1}
	neighbours[5] = Neighbour{X: 0, Y: -1}
	neighbours[6] = Neighbour{X: 1, Y: -1}
	neighbours[7] = Neighbour{X: 2, Y: -1}

	neighbours[8] = Neighbour{X: -2, Y: 0}
	neighbours[9] = Neighbour{X: -1, Y: 0}
	neighbours[10] = Neighbour{X: 1, Y: 0}
	neighbours[11] = Neighbour{X: 2, Y: 0}

	neighbours[12] = Neighbour{X: -2, Y: 1}
	neighbours[13] = Neighbour{X: -1, Y: 1}
	neighbours[14] = Neighbour{X: 0, Y: 1}
	neighbours[15] = Neighbour{X: 1, Y: 1}
	neighbours[16] = Neighbour{X: 2, Y: 1}

	neighbours[17] = Neighbour{X: -1, Y: 2}
	neighbours[18] = Neighbour{X: 0, Y: 2}
	neighbours[19] = Neighbour{X: 1, Y: 2}

	return neighbours
}

// RingNeighbours returns a ring of neigbours
func RingNeighbours() []Neighbour {
	neighbours := make([]Neighbour, 12)

	neighbours[0] = Neighbour{X: -1, Y: -2}
	neighbours[1] = Neighbour{X: 0, Y: -2}
	neighbours[2] = Neighbour{X: 1, Y: -2}

	neighbours[3] = Neighbour{X: -2, Y: -1}
	neighbours[4] = Neighbour{X: 2, Y: -1}

	neighbours[5] = Neighbour{X: -2, Y: 0}
	neighbours[6] = Neighbour{X: 2, Y: 0}

	neighbours[7] = Neighbour{X: -2, Y: 1}
	neighbours[8] = Neighbour{X: 2, Y: 1}

	neighbours[9] = Neighbour{X: -1, Y: 2}
	neighbours[10] = Neighbour{X: 0, Y: 2}
	neighbours[11] = Neighbour{X: 1, Y: 2}

	return neighbours
}

// GameOfLifeRules returns the RuleMap for Conway's Game Of Life
func GameOfLifeRules() RuleMap {
	rm := make(RuleMap)

	// alive (black) cell remains alive if it has 2 or 3 alive nbs
	rm[0] = []Rule{
		{
			Lower:          2,
			Upper:          3,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     EightNeighbours(),
		},
	}
	// dead (white) cell becomes alive if it has 3 alive nbs
	rm[1] = []Rule{
		{
			Lower:          3,
			Upper:          3,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     EightNeighbours(),
		},
	}

	return rm
}

// MazeRules returns the RuleMap for a maze creating effect
func MazeRules() RuleMap {
	rm := make(RuleMap)

	// alive (black) cell remains alive if it has 2 or 3 alive nbs
	rm[0] = []Rule{
		{
			Lower:          2,
			Upper:          4,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     EightNeighbours(),
		},
	}
	// dead (white) cell becomes alive if it has 3 alive nbs
	rm[1] = []Rule{
		{
			Lower:          3,
			Upper:          3,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     EightNeighbours(),
		},
	}

	return rm
}

// RockPaperScissors : rock crushes scissors, scissors cut paper, paper wraps rock
func RockPaperScissors(colorAmount int) RuleMap {
	rm := make(RuleMap)

	for i := 0; i < colorAmount; i++ {
		rules := []Rule{{
			Lower:          8,
			Upper:          20,
			NeighbourColor: uint8((i + 1) % colorAmount),
			NewColorIf:     uint8((i + 1) % colorAmount),
			NewColorElse:   uint8(i),
			Neighbours:     CircleNeighbours(),
		}}
		rm[uint8(i)] = rules
	}

	return rm
}

// Crystalisation : crystal forming structures
func Crystalisation(colorAmount int) RuleMap {
	rm := make(RuleMap)

	for i := 0; i < colorAmount; i++ {
		rules := []Rule{{
			Lower:          7,
			Upper:          8,
			NeighbourColor: uint8((i + 1) % colorAmount),
			NewColorIf:     uint8((i + 1) % colorAmount),
			NewColorElse:   uint8(i),
			Neighbours:     CircleNeighbours(),
		}}
		rm[uint8(i)] = rules
	}

	return rm
}

// Custom ...
func Custom() RuleMap {
	rm := make(RuleMap)

	// alive (black) cell remains alive if it has 2 or 3 alive nbs
	rm[0] = []Rule{
		{
			Lower:          2,
			Upper:          2,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     NeighboursLeft(),
		},
		{
			Lower:          3,
			Upper:          5,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     TwelveNeighbours(),
		},
	}
	// dead (white) cell becomes alive if it has 3 alive nbs
	rm[1] = []Rule{
		{
			Lower:          2,
			Upper:          5,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     TwelveNeighbours(),
		},
		{
			Lower:          2,
			Upper:          2,
			NeighbourColor: 0,
			NewColorIf:     0,
			NewColorElse:   1,
			Neighbours:     NeighboursRight(),
		},
	}

	return rm
}

// AvgRules will average to the nearby most occurring neighbour color
func AvgRules(colorAmount int) RuleMap {
	rm := make(RuleMap)

	for i := 0; i < colorAmount; i++ {
		rules := make([]Rule, colorAmount-1)
		ruleno := 0
		for j := i - 1; j != i; j = (j + 1) % colorAmount {
			rules[ruleno] = Rule{
				Lower:          5,
				Upper:          8,
				NeighbourColor: uint8(j),
				NewColorIf:     uint8(j),
				NewColorElse:   uint8(i),
				Neighbours:     EightNeighbours(),
			}
			ruleno++
		}
		rm[uint8(i)] = rules
	}

	return rm
}

func (nb Neighbour) checkBounds(img *image.Paletted, x, y int) bool {
	if (x+nb.X < 0) || (x+nb.X >= img.Rect.Dx()) {
		return false
	}
	if (y+nb.Y < 0) || (y+nb.Y >= img.Rect.Dy()) {
		return false
	}
	return true
}

func countNeighbours(img *image.Paletted, color uint8, x, y int, nbs []Neighbour) (count int) {
	count = 0
	for _, nb := range nbs {
		if nb.checkBounds(img, x, y) && img.ColorIndexAt(x+nb.X, y+nb.Y) == color {
			count++
		}
	}
	return
}

func (rule Rule) apply(img *image.Paletted, x, y int) uint8 {
	count := countNeighbours(img, rule.NeighbourColor, x, y, rule.Neighbours)
	if rule.Lower <= count && count <= rule.Upper {
		return rule.NewColorIf
	}
	return rule.NewColorElse

}

// ApplyRules applies the submitted rulemap to the inputted image
func (rm RuleMap) ApplyRules(img *image.Paletted) *image.Paletted {
	// create new image of some size and with the same palette
	newImg := image.NewPaletted(image.Rectangle{image.Pt(0, 0), image.Pt(img.Rect.Dx(), img.Rect.Dy())}, img.Palette)

	Xs := make([]int, img.Rect.Dx())
	for i := 0; i < len(Xs); i++ {
		Xs[i] = i
	}

	XSlices := needle.ChunkSlice(Xs, runtime.GOMAXPROCS(0))

	wg := sync.WaitGroup{}

	for _, XSlice := range XSlices {
		wg.Add(1)
		go func(Xs []int) {
			// for every pixel, apply the rules that concern its color
			for _, x := range Xs {
				for y := 0; y < img.Rect.Dy(); y++ {
					for _, rule := range rm[img.ColorIndexAt(x, y)] {
						newImg.SetColorIndex(x, y, rule.apply(img, x, y))
					}
				}
			}
			wg.Done()
		}(XSlice)
	}

	wg.Wait()

	return newImg
}

// PlayGame goes through an amount of iterations of a game based on the given rulemap
func (rm RuleMap) PlayGame(img *image.Paletted, iterations int, outputFile string, delay int) {
	var lastFrame *image.Paletted
	frames := make([]*image.Paletted, iterations+1)

	frames[0] = img
	lastFrame = frames[0]

	pb := pacebar.Pacebar{
		Work: iterations,
		Name: "GameOfColor",
	}

	for i := 0; i < iterations; i++ {
		lastFrame = rm.ApplyRules(lastFrame)
		frames[i+1] = lastFrame
		pb.Done(1)
	}

	gifeo.EncodeGIF(frames, outputFile, delay)
}
