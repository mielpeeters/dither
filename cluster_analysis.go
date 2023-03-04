package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type KMeansProblem struct {
	kMeans         PointSet //The estimated cluster centers (at this step)
	points         PointSet //The set of Points with kardinality n to subset into k clusters
	k              int      //The dimension of this k-means problem (amount of clusters)
	iterationStep  int      //The iteration step
	clusters       []PointSet
	maxDist        float64 //Maximum distance within the hyperbox containing all points
	distanceMetric func(pnt1, pnt2 Point) float64
	currentError   float64
}

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

// getContainingClusterIndex returns the index of the cluster that holds the point
//
// returns -1, -1 if none hold the point
func (KM *KMeansProblem) getContainingClusterIndex(point Point) (int, int) {
	// returns -1 if no cluster was found which contains the given point
	for clusterIndex, cluster := range KM.clusters {
		contains, containIndex := (&cluster).contains(point)
		if contains {
			return clusterIndex, containIndex
		}
	}
	return -1, -1
}

// assignment performs the assignment step of the KMeans algorithm: assigning points to clusters.
func (KM *KMeansProblem) assignment() {
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	for pointIndex, point := range KM.points.Points {
		wg.Add(1)
		go func(pointIndex int, point Point) {
			bestIndex := ClosestMeanIndex(KM, pointIndex)

			//only do something if that cluster doesn't already contain this point
			var newContains bool = false
			newContains, _ = (&KM.clusters[bestIndex]).contains(point)

			if !newContains {
				currentContainIndex, containingInternal := KM.getContainingClusterIndex(point)

				lock.Lock()
				if currentContainIndex > -1 {
					//delete from containing cluster
					KM.clusters[currentContainIndex].remove(containingInternal)
				}

				//add to better cluster
				KM.clusters[bestIndex].Points = append(KM.clusters[bestIndex].Points, point)
				lock.Unlock()
			}
			wg.Done()
		}(pointIndex, point)
	}
	wg.Wait()
}

// update performs the update step in the KMeans algorithm: update the means to be the mean of their clusters
func (KM *KMeansProblem) update() float64 {
	// calculating the means
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	changes := []float64{}

	for clusterId := range KM.clusters {
		wg.Add(1)
		go func(clusterId int) {
			old := KM.kMeans.Points[clusterId]
			mean := (&KM.clusters[clusterId]).mean()
			if len(mean.Coordinates) == 0 {
				KM.kMeans.Points[clusterId] = createRandomStart(KM.points, 1).Points[0] //bad choice, try another one
			} else {
				KM.kMeans.Points[clusterId] = mean
			}
			change := redMeanDistance(old, KM.kMeans.Points[clusterId])
			lock.Lock()
			changes = append(changes, change)
			lock.Unlock()
			wg.Done()
		}(clusterId)
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

func createRandomStart(points PointSet, k int) PointSet {
	//Get bounds so that the random starting points will at least lie in a reasonable region
	bounds := (&points).LowerAndUpperBounds()

	returnValue := PointSet{
		[]Point{},
	}

	if len(bounds) < 1 {
		return returnValue
	}

	//Re-seed for actual randomness
	rand.Seed(time.Now().UnixNano())

	//set the dimension
	dim := points.Points[0].Dimension()

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

func createKMeansProblem(points PointSet, k int, distanceMetric func(pnt1, pnt2 Point) float64) KMeansProblem {
	kMeans := createRandomStart(points, k)

	//Craete the initial clusters, consisting of just the random means in k different PointSets
	initClusters := []PointSet{}
	for i := 0; i < k; i++ {
		initClusters = append(initClusters, PointSet{})
	}

	bounds := (&points).LowerAndUpperBounds()

	// point 1 has the lowest coordinate value of all points
	// point 2 has the highest coordinate value of all points
	point1 := Point{}
	point2 := Point{}

	for dim := 0; dim < points.Points[0].Dimension(); dim++ {
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
