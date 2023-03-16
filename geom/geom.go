package geom

import (
	"math"
	"sort"
)

// Point is a collection of coordinates, with an identifier
type Point struct {
	Coordinates []float32
	ID          int
}

// PointSet implements a slice of points
type PointSet struct {
	Points []Point
}

// Bounds is a struct for lower - upper bounds
type Bounds struct {
	Lower float32
	Upper float32
}

// Dimension returns the dimension of the space this point lives in
func (p *Point) Dimension() int {
	return len(p.Coordinates)
}

// Equals determines whether or not two points are the same, including their IDs
func (p *Point) Equals(point Point) bool {
	if p.Dimension() != point.Dimension() { //check equality of Dimension
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

// Kardinality returns the kardinality doesn't it
func (ps *PointSet) Kardinality() int {
	return len(ps.Points)
}

// Contains determines wheter or not, and where, the given Point
// resides in the PointSet
func (ps *PointSet) Contains(point Point) (bool, int) {
	//check all points in ps for equality to point
	for i, pnt := range ps.Points {
		if pnt.Equals(point) {
			return true, i
		}
	}
	return false, -1
}

// ChunkPoints splits the given PointSet in chunks of size chunkSize.
// The last chunk might be smaller if chunkSize is not a factor of len(PointSet.Points)
func (ps *PointSet) ChunkPoints(chunkSize int) [][]Point {
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

// Remove removes the element at index from the PointSet, if a valid index is supplied
func (ps *PointSet) Remove(index int) {
	if index >= len(ps.Points) {
		return
	}

	ps.Points[index] = ps.Points[len(ps.Points)-1]

	ps.Points = ps.Points[:len(ps.Points)-1]
}

// Mean calculates the mean Point of all the Points in PointSet.
func (ps *PointSet) Mean() Point {
	meanCoords := []float32{}

	if len(ps.Points) == 0 {
		return Point{[]float32{}, 0}
	}
	for dim := 0; dim < ps.Points[0].Dimension(); dim++ {
		meanCoords = append(meanCoords, 0.0)
	}

	for _, point := range ps.Points { // for each point
		for i := 0; i < point.Dimension(); i++ { //for each dimension
			meanCoords[i] += point.Coordinates[i] / float32(len(ps.Points))
		}
	}
	meanPoint := Point{
		meanCoords,
		0,
	}

	return meanPoint
}

// LowerAndUpperBounds returns Bounds for each dimension of the space wherein the
// Points of PointSet live.
// Returned value is a collection of lower and upper bounds, one Bounds for each dimension!
func (ps *PointSet) LowerAndUpperBounds() []Bounds {

	bounds := []Bounds{}

	if len(ps.Points) < 1 {
		return bounds
	}

	dim := ps.Points[0].Dimension()

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

// SortByAxis sorts all Points in the PointSet in ascending order of certain provided axis numbers coordinate.
func (ps *PointSet) SortByAxis(axis int) {
	sort.Slice(ps.Points, func(i int, j int) bool {
		return ps.Points[i].Coordinates[axis] < ps.Points[j].Coordinates[axis]
	})
}

// BranchByMedian splits a PointSet in a left and a right PointSet, and also returns the Point that splits the two.
// This is done by sorting on one axis (see (*PointSet).SortByAxis())
func (ps *PointSet) BranchByMedian(axis int) (PointSet, PointSet, Point) {
	ps.SortByAxis(axis)

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

// EuclidianDistance returns the euclidian distance of two points
func EuclidianDistance(pnt1, pnt2 Point) float64 {
	var dist float64
	for index := range pnt1.Coordinates {
		dist += math.Pow(float64(pnt1.Coordinates[index]-pnt2.Coordinates[index]), 2)
	}

	return dist
}

// RedMeanDistance returns the red mean distance of two color points,
// thus only works with the first 3 dimensions of the points
func RedMeanDistance(pnt1, pnt2 Point) float64 {
	// only to use with colors!
	redMean := (pnt1.Coordinates[0] + pnt2.Coordinates[0]) / 2

	output := float64(2+redMean/256) * math.Pow(float64(pnt1.Coordinates[0]-pnt2.Coordinates[0]), 2)

	output += 4 * math.Pow(float64(pnt1.Coordinates[1]-pnt2.Coordinates[1]), 2)

	output += float64(2+(255-redMean)/256) * math.Pow(float64(pnt1.Coordinates[2]-pnt2.Coordinates[2]), 2)

	return output
}
