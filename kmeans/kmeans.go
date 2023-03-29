package kmeans

import (
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/mielpeeters/dither/geom"
)

// Clustering is a K Means clustering struct
type Clustering struct {
	KMeans         geom.PointSet //The estimated cluster centers (at this step)
	points         geom.PointSet //The set of Points with kardinality n to subset into k clusters
	k              int           //The dimension of this k-means problem (amount of clusters)
	Clusters       []geom.PointSet
	maxDist        float64 //Maximum distance within the hyperbox containing all points
	distanceMetric func(pnt1, pnt2 *geom.Point) float64
	// batch          []*geom.Point
}

var maxBatchSize = 30000
var iterationLimit = 100

// ClosestMeanIndex returns the index within the KM.kMeans slice
// of that mean which is closest to the given point, by index pointIndex (stored in KM.points)
func ClosestMeanIndex(KM *Clustering, pointIndex int) int {
	var minDist float64
	var bestIndex int
	for meanIndex := range KM.KMeans.Points { //check all means

		dist := KM.distanceMetric(&KM.points.Points[pointIndex], &KM.KMeans.Points[meanIndex])

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

// assign performs the assignment step of the KMeans algorithm: assigning points to clusters.
func (KM *Clustering) assign() {
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	workers := runtime.GOMAXPROCS(0)

	var pointChunks [][]geom.Point
	// try to divide amongst the amount of workers
	dividedAmount := int(math.Ceil(float64(len(KM.points.Points))) / float64(workers))

	var batchSize int
	if len(KM.points.Points) > maxBatchSize {
		batchSize = maxBatchSize / workers
	} else {
		batchSize = dividedAmount
	}

	pointChunks = KM.points.ChunkPointsMiniBatch(workers, batchSize)

	// KM.batch = make([]*geom.Point, 0)
	// for i := range pointChunks {
	// 	KM.batch = append(KM.batch, pointChunks[i]...)
	// }

	startIndex := 0

	// reset clusters
	KM.Clusters = make([]geom.PointSet, KM.k)

	// handle each chunk in parallel
	for _, points := range pointChunks {
		wg.Add(1)

		go func(points []geom.Point, startIndex int) {
			newClusters := make([]geom.PointSet, KM.k)

			for i, point := range points {
				bestIndex := ClosestMeanIndex(KM, startIndex+i)
				newClusters[bestIndex].Points = append(newClusters[bestIndex].Points, point)
			}

			lock.Lock()
			for cluster := range KM.Clusters {
				KM.Clusters[cluster].Points = append(KM.Clusters[cluster].Points, newClusters[cluster].Points...)
			}
			lock.Unlock()

			wg.Done()
		}(points, startIndex)

		startIndex += len(points)
	}
	wg.Wait()
}

// update performs the update step in the KMeans algorithm: update the means to be the mean of their clusters
func (KM *Clustering) update() float64 {
	// calculating the means
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	changes := []float64{}

	for clusterID := range KM.Clusters {
		wg.Add(1)
		go func(clusterID int) {
			old := KM.KMeans.Points[clusterID]
			mean := (&KM.Clusters[clusterID]).Mean()
			if len(mean.Coordinates) == 0 {
				KM.KMeans.Points[clusterID] = createRandomStart(KM.points, 1).Points[0] //bad choice, try another one
			} else {
				KM.KMeans.Points[clusterID] = mean
			}
			change := KM.distanceMetric(&old, &KM.KMeans.Points[clusterID])
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

// TotalDist returns the total distance from points to their assigned cluster mean
func (KM *Clustering) TotalDist() float64 {

	var sum float64

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	for meanIndex := range KM.KMeans.Points { // iterate over all means
		wg.Add(1)
		go func(points []geom.Point, meanIndex int) {
			localSum := 0.0
			for pointIndex := range points {
				localSum += KM.distanceMetric(&KM.KMeans.Points[meanIndex], &KM.Clusters[meanIndex].Points[pointIndex])
			}

			mutex.Lock()
			sum += localSum
			mutex.Unlock()

			wg.Done()
		}(KM.Clusters[meanIndex].Points, meanIndex)
	}

	wg.Wait()
	return sum
}

// iterate performs one iteration of the KMeans algorithm
//
// Returns:
//   - whether accuracy was met, as a bool
//   - the achieved change, maxChange / KM.maxDist, as a percentage (float)
func (KM *Clustering) iterate(accuracy float64) bool {
	KM.assign()
	maxChange := KM.update()

	return (maxChange * 100 / KM.maxDist) < accuracy
}

func createRandomStart(points geom.PointSet, k int) geom.PointSet {
	//Get bounds so that the random starting points will at least lie in a reasonable region
	bounds := (&points).LowerAndUpperBounds()

	returnValue := geom.PointSet{
		Points: []geom.Point{},
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
		var currentPoint geom.Point
		for dimNum := 0; dimNum < dim; dimNum++ {
			low = bounds[dimNum].Lower                                                                //lower bound for this coordinate number
			upp = bounds[dimNum].Upper                                                                //upper bound for this coordinate number
			currentPoint.Coordinates = append(currentPoint.Coordinates, rand.Float32()*(upp-low)+low) //random value between corr. bounds
		}
		returnValue.Points = append(returnValue.Points, currentPoint) // add the fully random point to the geom.PointSet
	}

	return returnValue
}

// CreateKMeansProblem generates a new k-means clustering problem.
//
// points is the PointSet that contains the clusters that are to be found. k is the estimated amount of clusters.
// distanceMetric is the function to be used for determining "closeness"
func CreateKMeansProblem(points geom.PointSet, k int, distanceMetric func(pnt1, pnt2 *geom.Point) float64) Clustering {
	kMeans := createRandomStart(points, k)

	//Craete the initial clusters, consisting of just the random means in k different geom.PointSets
	initClusters := make([]geom.PointSet, k)

	bounds := (&points).LowerAndUpperBounds()

	// point 1 has the lowest coordinate value of all points
	// point 2 has the highest coordinate value of all points
	point1 := geom.Point{}
	point2 := geom.Point{}

	for dim := 0; dim < points.Points[0].Dimension(); dim++ {
		point1.Coordinates = append(point1.Coordinates, bounds[dim].Lower)
		point2.Coordinates = append(point2.Coordinates, bounds[dim].Upper)
	}

	maxDist := distanceMetric(&point1, &point2)

	returnValue := Clustering{
		kMeans,
		points,
		k,
		initClusters,
		maxDist,
		distanceMetric,
	}

	return returnValue
}

// Cluster performs the clustering algorithm, with specified parameters for accuracy
//
//   - accuracy: the amount of relative change below which the algorithm is considered to have converged
//   - consecutiveTimes: the amount of times the accuracy has to be met consecutively for convergence
func (KM *Clustering) Cluster(accuracy float64, consecutiveTimes int) {
	var done bool
	var consecutiveDone int

	count := 0

	for consecutiveDone < consecutiveTimes && count < iterationLimit {
		count++
		done = KM.iterate(accuracy)
		if done {
			consecutiveDone++
		} else {
			consecutiveDone = 0
		}
	}
}
