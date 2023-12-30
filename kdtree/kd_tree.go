package kdtree

import (
	"fmt"
	"math"
	"time"

	"github.com/mielpeeters/dither/geom"
)

// KDTree is a kd tree struct
type KDTree struct {
	Root     *Node
	Lookup   map[int]geom.Point
	BestDist float64
}

// Node is a node struct for within a KD tree
type Node struct {
	PointValue []geom.Point
	Left       *Node
	Right      *Node
	Parrent    *Node
}

func (node *Node) isLeafNode() bool {
	return node.Left == nil
}

func (node *Node) isRootNode() bool {
	return node.Parrent == nil
}

func meanCutAlgorithm(points geom.PointSet) geom.PointSet {
	var returnSet geom.PointSet

	tree := generateKDTreeFromPoints(points, 5)

	var leafs [][]geom.Point
	tree.Root.leafs(&leafs)

	for _, leaf := range leafs {
		set := geom.PointSet{
			Points: leaf,
		}

		// find the mean of the leaf\
		mean := set.Mean()
		returnSet.Points = append(returnSet.Points, mean)
	}

	return returnSet
}

func generateKDTreeFromPoints(points geom.PointSet, depth int) KDTree {
	var kd KDTree

	nmbAxis := len(points.Points[0].Coordinates)

	root := generateKDNodeFromPoints(points, 0, nmbAxis, depth)

	kd.Root = root

	kd.BestDist = -1.0

	return kd
}

// Recursively create a KDTree from a set of points, on the given start axis
// and with the given number of axis
//
// # Arguments:
//   - points: the set of points to generate the node from
//   - axis: the axis to split the node on
//   - nmbAxis: the number of axis in total
//   - depth: the depth to grow the tree to
func generateKDNodeFromPoints(points geom.PointSet, axis int, nmbAxis int, depth int) *Node {
	if depth == 0 {
		return &Node{
			points.Points,
			nil,
			nil,
			nil,
		}
	}
	// generate a left and a right pointset
	leftSet, rightSet, pivot := points.BranchByMedian(axis)

	thisNode := Node{
		[]geom.Point{pivot},
		nil,
		nil,
		nil,
	}

	var leftNode *Node
	var rightNode *Node

	if len(leftSet.Points) > 0 {
		// create the current node
		leftNode = generateKDNodeFromPoints(leftSet, (axis+1)%nmbAxis, nmbAxis, depth-1)
		leftNode.Parrent = &thisNode
		thisNode.Left = leftNode

		if len(rightSet.Points) > 0 {
			rightNode = generateKDNodeFromPoints(rightSet, (axis+1)%nmbAxis, nmbAxis, depth-1)
			rightNode.Parrent = &thisNode
			thisNode.Right = rightNode
		}

	} else {
		thisNode.Left = nil
		thisNode.Right = nil
	}

	return &thisNode
}

func (kd *KDTree) print() {
	fmt.Println("")
	fmt.Println("")
	fmt.Println("___________THIS IS A KD-TREE______________")
	fmt.Println("")
	kd.Root.print(0)
	fmt.Println("")
	fmt.Println("___________THIS WAS A KD-TREE_____________")
	fmt.Println("")
	fmt.Println("")
}

func (node *Node) print(level int) {
	var space string
	curLev := level
	for curLev > 0 {
		curLev--
		space += "|    "
	}

	fmt.Println(space, "* [STARTNODE]", node.PointValue, "*")
	fmt.Println(space, "* PARENT:", node.Parrent, "*")

	if node.Left != nil {
		fmt.Println(space, "left:")
		(node.Left).print(level + 1)
		if node.Right != nil {
			fmt.Println(space, "right:")
			(node.Right).print(level + 1)
		} else {
			fmt.Println(space, "no right node...")
		}
	}

	fmt.Println(space, "* [ENDNODE] *")
}

// goDownOneLevel returns the node that is one level down from the current node,
// and closest to the given point
func (node *Node) goDownOneLevel(point geom.Point, level int) (*Node, bool) {
	var returnNode *Node
	var returnCode bool
	if point.Coordinates[level] < node.PointValue[0].Coordinates[level] {
		if node.Left != nil {
			returnNode = node.Left
			returnCode = true
		}
	} else {
		if node.Right != nil {
			returnNode = node.Right
			returnCode = true
		} else if node.Left != nil {
			returnNode = node.Left
			returnCode = true
		}

	}
	return returnNode, returnCode
}

func (node *Node) goUpOneLevel() *Node {
	returnNode := node.Parrent

	return returnNode
}

func (node *Node) leafs(leaf_vals *[][]geom.Point) {
	if len(node.PointValue) > 1 {
		// this is a leaf node
		*leaf_vals = append(*leaf_vals, node.PointValue)
		return
	}

	if node.Left != nil {
		node.Left.leafs(leaf_vals)
	} else if node.Right != nil {
		node.Right.leafs(leaf_vals)
	}
}

func (kd *KDTree) findNearestNeighborTo(point geom.Point, distanceMetricFunction func(geom.Point, geom.Point) float64, nmbAxis int) (geom.Point, float64) {
	var currentLevel int
	var currentBest geom.Point
	var currentNode *Node
	var lastNode *Node
	var exists bool
	var currentBestDist float64

	// first, traverse the entire tree until we reach a leafnode
	currentNode = kd.Root

	for !currentNode.isLeafNode() {
		lastNode = currentNode
		currentNode, exists = currentNode.goDownOneLevel(point, currentLevel%nmbAxis)

		if !exists {
			currentNode = lastNode
			break
		} else {
			currentLevel++
		}
	}

	// store the current best distance
	currentBest = currentNode.PointValue[0]
	currentBestDist = distanceMetricFunction(currentBest, point)

	if kd.BestDist == -1 || currentBestDist < kd.BestDist {
		kd.BestDist = currentBestDist
	}

	// now, go up the tree again, until we reach the rootnode again
	// each time, check if the other branch might contain a better neighbor
	//	and if the current node might be closer itself

	for !currentNode.isRootNode() {
		// go up one level, to the parent node
		lastNode = currentNode
		currentNode = (currentNode).goUpOneLevel()
		currentLevel--

		var hyperplanedist float64

		if currentLevel < 0 {
			break
		}

		hyperplanedist = math.Pow(float64(point.Coordinates[currentLevel%nmbAxis]-currentNode.PointValue[0].Coordinates[currentLevel%nmbAxis]), 2)

		if kd.BestDist > hyperplanedist {
			// the hypersphere intersects with the hyperplane
			// thus the other branch side could contain a better neighbor!

			// create a new kdtree, being the other branch
			var newKd KDTree
			var newRoot Node
			// use the other branch!
			if currentNode.Left == lastNode {
				// came from Left branch

				if currentNode.Right != nil {
					newRoot = *currentNode.Right
				} else {
					// other side is empty!

					goto noIntersect
				}
			} else if currentNode.Right == lastNode {
				// came from Right branch
				newRoot = *currentNode.Left
			} else {
				fmt.Println("!!!!!!!something went wrong in the left/right thing!!!!!!!!!")
				time.Sleep(time.Second)
			}
			newRoot.Parrent = nil // make it a root node...
			newKd.Root = &newRoot

			otherBest, otherBestDist := newKd.findNearestNeighborTo(point, distanceMetricFunction, nmbAxis)
			if otherBestDist < currentBestDist {
				currentBest = otherBest
				currentBestDist = otherBestDist
				kd.BestDist = currentBestDist
			}
		}
	noIntersect:
		// lastly, check whether the currentNode itself (which is the parent of the last one, possibly a root!) is closer
		otherBestDist := distanceMetricFunction(currentNode.PointValue[0], point)
		if otherBestDist < currentBestDist {
			currentBest = currentNode.PointValue[0]
			currentBestDist = otherBestDist
			kd.BestDist = currentBestDist
		}
	}

	return currentBest, currentBestDist
}
