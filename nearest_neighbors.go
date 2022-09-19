package main

import "sort"

type Point struct {
	Coordinates []float64
	Id          int
}

func (p Point) Dimension() int {
	return len(p.Coordinates)
}

func (p Point) equals(point Point) bool {
	if p.Dimension() != point.Dimension() { //check equality of Dimension
		return false
	}
	if p.Id != point.Id {
		return false
	}
	for i := range p.Coordinates {
		if p.Coordinates[i] != point.Coordinates[i] {
			return false
		}
	}

	return true
}

type PointSet struct {
	Points []Point
}

func (ps PointSet) Kardinality() int {
	return len(ps.Points)
}

func (ps PointSet) contains(point Point) (bool, int) {
	for i := range ps.Points {
		if ps.Points[i].equals(point) { //check all points in ps for equality to point
			return true, i
		}
	}
	return false, -1
}

func (ps PointSet) remove(index int) PointSet {
	if index >= len(ps.Points) {
		return ps
	}

	ps.Points[index] = ps.Points[len(ps.Points)-1]

	returnValue := PointSet{
		ps.Points[:len(ps.Points)-1],
	}
	return returnValue
}

func (ps PointSet) mean() Point {
	meanCoords := []float64{}

	if len(ps.Points) == 0 {
		return Point{[]float64{}, 0}
	} else {
		for dim := 0; dim < ps.Points[0].Dimension(); dim++ {
			meanCoords = append(meanCoords, 0.0)
		}
	}

	for _, point := range ps.Points { // for each point
		for i := 0; i < point.Dimension(); i++ { //for each dimension
			meanCoords[i] += point.Coordinates[i] / float64(len(ps.Points))
		}
	}
	meanPoint := Point{
		meanCoords,
		0,
	}

	return meanPoint
}

func (ps PointSet) LowerAndUpperBounds() []Bounds {
	// return value is a collection of lower and upper bounds, for each dimension!

	bounds := []Bounds{}

	if len(ps.Points) < 1 {
		return bounds
	}

	dim := ps.Points[0].Dimension()

	var currentLower float64
	var currentUpper float64

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

func (ps PointSet) sortByAxis(axis int) {
	sort.Slice(ps.Points, func(i int, j int) bool {
		if ps.Points[i].Coordinates[axis] < ps.Points[j].Coordinates[axis] {
			return true
		} else {
			return false
		}
	})
}

func (ps PointSet) branchByMedian(axis int) (PointSet, PointSet, Point) {
	ps.sortByAxis(axis)

	medianIndex := len(ps.Points) / 2

	left := ps.Points[:medianIndex]
	right := ps.Points[medianIndex+1:]

	leftSet := PointSet{
		left,
	}
	rightSet := PointSet{
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
