package main

import (
	"sort"
)

// Point point
type Point struct {
	Coordinates []float32
	ID          int
}

func (p *Point) dimension() int {
	return len(p.Coordinates)
}

func (p *Point) equals(point Point) bool {
	if p.dimension() != point.dimension() { //check equality of Dimension
		return false
	}
	if p.ID != point.ID {
		return false
	}
	for i := range p.Coordinates {
		if p.Coordinates[i] != point.Coordinates[i] {
			return false
		}
	}

	return true
}

type pointSet struct {
	Points []Point
}

func (ps *pointSet) Kardinality() int {
	return len(ps.Points)
}

func (ps *pointSet) contains(point Point) (bool, int) {
	// wg := sync.WaitGroup{}

	for i, pnt := range ps.Points {
		if pnt.equals(point) { //check all points in ps for equality to point
			return true, i
		}
	}
	return false, -1
}

func (ps *pointSet) chunkPoints(chunkSize int) [][]Point {
	var chunks [][]Point
	for i := 0; i < len(ps.Points); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// ps.Points capacity
		if end > len(ps.Points) {
			end = len(ps.Points)
		}

		chunks = append(chunks, ps.Points[i:end])
	}

	return chunks
}

// removes the element at index from the PointSet, if alowed
func (ps *pointSet) remove(index int) {
	if index >= len(ps.Points) {
		return
	}

	ps.Points[index] = ps.Points[len(ps.Points)-1]

	ps.Points = ps.Points[:len(ps.Points)-1]
}

func (ps *pointSet) mean() Point {
	meanCoords := []float32{}

	if len(ps.Points) == 0 {
		return Point{[]float32{}, 0}
	}
	for dim := 0; dim < ps.Points[0].dimension(); dim++ {
		meanCoords = append(meanCoords, 0.0)
	}

	for _, point := range ps.Points { // for each point
		for i := 0; i < point.dimension(); i++ { //for each dimension
			meanCoords[i] += point.Coordinates[i] / float32(len(ps.Points))
		}
	}
	meanPoint := Point{
		meanCoords,
		0,
	}

	return meanPoint
}

func (ps *pointSet) LowerAndUpperBounds() []Bounds {
	// return value is a collection of lower and upper bounds, for each dimension!

	bounds := []Bounds{}

	if len(ps.Points) < 1 {
		return bounds
	}

	dim := ps.Points[0].dimension()

	var currentLower float32
	var currentUpper float32

	for coordNum := 0; coordNum < dim; coordNum++ {
		currentLower = ps.Points[0].Coordinates[coordNum]
		currentUpper = currentLower
		for _, point := range ps.Points {
			if point.Coordinates[coordNum] < currentLower {
				currentLower = point.Coordinates[coordNum]
			}

			if point.Coordinates[coordNum] > currentUpper {
				currentUpper = point.Coordinates[coordNum]
			}
		}
		currentBounds := Bounds{
			currentLower,
			currentUpper,
		}
		bounds = append(bounds, currentBounds)
	}

	return bounds
}

func (ps *pointSet) sortByAxis(axis int) {
	sort.Slice(ps.Points, func(i int, j int) bool {
		if ps.Points[i].Coordinates[axis] < ps.Points[j].Coordinates[axis] {
			return true
		}
		return false

	})
}

func (ps *pointSet) branchByMedian(axis int) (pointSet, pointSet, Point) {
	ps.sortByAxis(axis)

	medianIndex := len(ps.Points) / 2

	left := ps.Points[:medianIndex]
	right := ps.Points[medianIndex+1:]

	leftSet := pointSet{
		left,
	}
	rightSet := pointSet{
		right,
	}

	return leftSet, rightSet, ps.Points[medianIndex]
}

func findNearestNeighbor(neighbors []Point, point Point, distanceMetricFunction func(Point, Point) float64) Point {
	var bestOption Point
	var bestDistance float64
	var distance float64

	firstLoop := true

	for _, neighbor := range neighbors {
		distance = distanceMetricFunction(neighbor, point)

		if firstLoop || distance < bestDistance {
			firstLoop = false
			bestDistance = distance
			bestOption = neighbor
		}
	}

	return bestOption
}
