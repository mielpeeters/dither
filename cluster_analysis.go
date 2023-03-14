package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// KMeansProblem is a K Means clustering Problem struct
type KMeansProblem struct {
	kMeans         pointSet //The estimated cluster centers (at this step)
	points         pointSet //The set of Points with kardinality n to subset into k clusters
	k              int      //The dimension of this k-means problem (amount of clusters)
	iterationStep  int      //The iteration step
	clusters       []pointSet
	maxDist        float64 //Maximum distance within the hyperbox containing all points
	distanceMetric func(pnt1, pnt2 Point) float64
	currentError   float64
}

// Bounds is a struct for lower - upper bounds
type Bounds struct {
	lower float32
	upper float32
}

// ClosestMeanIndex returns the index within the KM.kMeans slice
// of that mean which is closest to the given point, by index pointIndex (stored in KM.points)
func ClosestMeanIndex(KM *KMeansProblem, pointIndex int) int {
	var minDist float64
	var bestIndex int
	for meanIndex := range KM.kMeans.Points { //check all means

		dist := KM.distanceMetric(KM.points.Points[pointIndex], KM.kMeans.Points[meanIndex])

		if meanIndex == 0 {
			minDist = dist
			bestIndex = meanIndex
		} else if dist < minDist {
			minDist = dist
			bestIndex = meanIndex
		}
	}

	return bestIndex
}

// assignment performs the assignment step of the KMeans algorithm: assigning points to clusters.
func (KM *KMeansProblem) assignment() {
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	// try to divide in 8 chunks, 9 also possible
	pointChunks := KM.points.chunkPoints(len(KM.points.Points) / 16)

	startIndex := 0

	// reset clusters
	KM.clusters = make([]pointSet, KM.k)

	// handle each chunk in parallel
	for _, points := range pointChunks {
		wg.Add(1)

		go func(points []Point, startIndex int) {
			newClusters := make([]pointSet, KM.k)

			for i, point := range points {
				bestIndex := ClosestMeanIndex(KM, startIndex+i)
				newClusters[bestIndex].Points = append(newClusters[bestIndex].Points, point)
			}

			lock.Lock()
			for cluster := range KM.clusters {
				KM.clusters[cluster].Points = append(KM.clusters[cluster].Points, newClusters[cluster].Points...)
			}
			lock.Unlock()

			wg.Done()
		}(points, startIndex)

		startIndex += len(points)
	}
	wg.Wait()
}

// update performs the update step in the KMeans algorithm: update the means to be the mean of their clusters
func (KM *KMeansProblem) update() float64 {
	// calculating the means
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	changes := []float64{}

	for clusterID := range KM.clusters {
		wg.Add(1)
		go func(clusterID int) {
			old := KM.kMeans.Points[clusterID]
			mean := (&KM.clusters[clusterID]).mean()
			if len(mean.Coordinates) == 0 {
				KM.kMeans.Points[clusterID] = createRandomStart(KM.points, 1).Points[0] //bad choice, try another one
			} else {
				KM.kMeans.Points[clusterID] = mean
			}
			change := KM.distanceMetric(old, KM.kMeans.Points[clusterID])
			lock.Lock()
			changes = append(changes, change)
			lock.Unlock()
			wg.Done()
		}(clusterID)
	}
	wg.Wait()

	var max float64
	for i := range changes {
		max = math.Max(max, changes[i])
	}

	return max
}

// totalDist returns the total distance from points to their assigned cluster mean
func (KM *KMeansProblem) totalDist() float64 {

	var sum float64

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	for meanIndex := range KM.kMeans.Points { // iterate over all means
		wg.Add(1)
		go func(points []Point, meanIndex int) {
			localSum := 0.0
			for pointIndex := range points {
				localSum += KM.distanceMetric(KM.kMeans.Points[meanIndex], KM.clusters[meanIndex].Points[pointIndex])
			}

			mutex.Lock()
			sum += localSum
			mutex.Unlock()

			wg.Done()
		}(KM.clusters[meanIndex].Points, meanIndex)
	}

	wg.Wait()
	return sum
}

// iterate performs one iteration of the KMeans algorithm
//
// returns:
//   - whether accuracy was met, as a bool
//   - the achieved change, maxChange / KM.maxDist, as a percentage (float)
func (KM *KMeansProblem) iterate(accuracy float64) (bool, float64) {
	KM.assignment()
	maxChange := KM.update()

	KM.currentError = KM.totalDist()

	return (maxChange * 100 / KM.maxDist) < accuracy, maxChange * 100 / KM.maxDist
}

func createRandomStart(points pointSet, k int) pointSet {
	//Get bounds so that the random starting points will at least lie in a reasonable region
	bounds := (&points).LowerAndUpperBounds()

	returnValue := pointSet{
		[]Point{},
	}

	if len(bounds) < 1 {
		return returnValue
	}

	//Re-seed for actual randomness
	rand.Seed(time.Now().UnixNano())

	//set the dimension
	dim := points.Points[0].dimension()

	var low float32
	var upp float32
	for i := 0; i < k; i++ {
		var currentPoint Point
		for dimNum := 0; dimNum < dim; dimNum++ {
			low = bounds[dimNum].lower                                                                //lower bound for this coordinate number
			upp = bounds[dimNum].upper                                                                //upper bound for this coordinate number
			currentPoint.Coordinates = append(currentPoint.Coordinates, rand.Float32()*(upp-low)+low) //random value between corr. bounds
		}
		returnValue.Points = append(returnValue.Points, currentPoint) // add the fully random point to the PointSet
	}

	return returnValue
}

func createKMeansProblem(points pointSet, k int, distanceMetric func(pnt1, pnt2 Point) float64) KMeansProblem {
	kMeans := createRandomStart(points, k)

	//Craete the initial clusters, consisting of just the random means in k different PointSets
	initClusters := make([]pointSet, k)

	bounds := (&points).LowerAndUpperBounds()

	// point 1 has the lowest coordinate value of all points
	// point 2 has the highest coordinate value of all points
	point1 := Point{}
	point2 := Point{}

	for dim := 0; dim < points.Points[0].dimension(); dim++ {
		point1.Coordinates = append(point1.Coordinates, bounds[dim].lower)
		point2.Coordinates = append(point2.Coordinates, bounds[dim].upper)
	}

	maxDist := distanceMetric(point1, point2)

	returnValue := KMeansProblem{
		kMeans,
		points,
		k,
		0,
		initClusters,
		maxDist,
		distanceMetric,
		0,
	}

	return returnValue
}
