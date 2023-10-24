package nearneigh

import "github.com/mielpeeters/dither/geom"

func findNearestNeighbor(neighbors []geom.Point, point geom.Point, distanceMetricFunction func(geom.Point, geom.Point) float64) geom.Point {
	var bestOption geom.Point
	var bestDistance float64
	var distance float64 = -1.0

	for _, neighbor := range neighbors {
		distance = distanceMetricFunction(neighbor, point)

		if distance < 0 || distance < bestDistance {
			firstLoop = false
			bestDistance = distance
			bestOption = neighbor
		}
	}

	return bestOption
}
