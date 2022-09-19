package main

import (
	"fmt"
	"math"
	"time"
)

type KDTree struct {
	Root     *Node
	Lookup   map[int]Point
	BestDist float64
}

type Node struct {
	PointValue Point
	Left       *Node
	Right      *Node
	Parrent    *Node
}

// type Level struct {
// 	Pairs []Pair
// }

// type Pair struct {
// 	Left  *Node
// 	Right *Node
// }

const printing = false

func (node Node) isLeafNode() bool {
	return node.Left == nil
}

func (node Node) isRootNode() bool {
	return node.Parrent == nil
}

func generateKDTreeFromPoints(points PointSet, nmbAxis int) KDTree {
	var kd KDTree

	root := generateKDNodeFromPoints(points, 0, nmbAxis)

	kd.Root = root

	kd.BestDist = -1.0

	return kd
}

func generateKDNodeFromPoints(points PointSet, axis int, nmbAxis int) *Node {
	// generate a left and a right pointset
	leftSet, rightSet, pivot := points.branchByMedian(axis)

	thisNode := Node{
		pivot,
		nil,
		nil,
		nil,
	}

	var leftNode *Node
	var rightNode *Node

	if len(leftSet.Points) > 0 {
		// create the current node
		leftNode = generateKDNodeFromPoints(leftSet, (axis+1)%nmbAxis, nmbAxis)
		leftNode.Parrent = &thisNode
		thisNode.Left = leftNode

		if len(rightSet.Points) > 0 {
			rightNode = generateKDNodeFromPoints(rightSet, (axis+1)%nmbAxis, nmbAxis)
			rightNode.Parrent = &thisNode
			thisNode.Right = rightNode
		}

	} else {
		thisNode.Left = nil
		thisNode.Right = nil
	}

	return &thisNode
}

func (kd KDTree) print() {
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

func (node Node) print(level int) {
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

func (node Node) goDownOneLevel(point Point, level int) (*Node, bool) {
	var returnNode *Node
	var returnCode bool
	if point.Coordinates[level] < node.PointValue.Coordinates[level] {
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

func (node Node) goUpOneLevel() *Node {
	returnNode := node.Parrent

	return returnNode
}

func (kd KDTree) findNearestNeighborTo(point Point, distanceMetricFunction func(Point, Point) float64, nmbAxis int) (Point, float64) {
	var currentLevel int
	var currentBest Point
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
	currentBest = currentNode.PointValue
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

		hyperplanedist = math.Pow((point.Coordinates[currentLevel%nmbAxis] - currentNode.PointValue.Coordinates[currentLevel%nmbAxis]), 2)

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
		otherBestDist := distanceMetricFunction(currentNode.PointValue, point)
		if otherBestDist < currentBestDist {
			currentBest = currentNode.PointValue
			currentBestDist = otherBestDist
			kd.BestDist = currentBestDist
		}
	}

	return currentBest, currentBestDist
}
