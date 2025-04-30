package algorithms

import (
	"container/heap"
	"fmt"
	"math"
	"os"

	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
)

// Node represents a vertex in the heap with its current degree and index.
// index is maintained by the heap for update operations.
type Node struct {
	id    int     // vertex ID
	deg   float64 // current degree
	index int     // index in the heap
}

// MinHeap implements a min-heap of *Node based on deg.
type MinHeap []*Node

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].deg < h[j].deg }
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *MinHeap) Push(x interface{}) {
	n := x.(*Node)
	n.index = len(*h)
	*h = append(*h, n)
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// update adjusts the degree of node in the heap and reorders.
func (h *MinHeap) update(node *Node, newDeg float64) {
	node.deg = newDeg
	heap.Fix(h, node.index)
}

// loadGraphRR loads the graph, applies randomized response (RR) to the upper
// triangular adjacency (i < j), and returns an n x n integer matrix (0/1) indicating edges.
func loadGraphRR(filename string, noise bool, bidirectional bool, epsilon float64, n int) ([][]int, error) {
	// Load original graph
	graph, err := datastructures.NewGraph(filename, bidirectional)
	if err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	// processedList holds RR-processed neighbor lists per node
	processedList := make(map[int][]int, n)

	// First pass: apply RR to each node's adjacency list
	for i := 0; i < n; i++ {
		processedList[i] = randomizedResponse(epsilon, graph.AdjacencyList[i], n, i)
	}

	// Initialize n x n integer matrix, default 0
	matrix := make([][]int, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]int, n)
	}

	// Fill upper-triangular entries based on RR output
	for i, neighs := range processedList {
		for _, j := range neighs {
			// only consider upper triangle, skip invalid IDs or self-loops
			if j > i && j < n {
				matrix[i][j] = 1
			}
		}
	}

	// Enforce symmetry: copy upper to lower triangular
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if matrix[i][j] == 1 {
				matrix[j][i] = 1
			}
		}
	}

	return matrix, nil
}

func KCoreRR(n int, psi float64, epsilon float64, noise bool, graphFileName string, outputFileName string) {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	graph, err := loadGraphRR(graphFileName, noise, true, epsilon, n)

	if err != nil {
		fmt.Println("error loading graph")
		return
	}

	// initialize degrees and scale due to RR
	degree := make(map[int]float64, n)
	for u, _ := range graph {
		for j := 0; j < n; j++ {
			degree[u] += (float64(graph[u][j])*(math.Exp(epsilon)+1) - 1) / (math.Exp(epsilon) - 1)
		}
	}

	// create nodes and heap
	nodes := make(map[int]*Node, len(graph))
	hArr := make(MinHeap, 0, len(graph))
	for u, d := range degree {
		n := &Node{id: u, deg: d}
		nodes[u] = n
		hArr = append(hArr, n)
	}
	for i, n := range hArr {
		n.index = i
	}
	heap.Init(&hArr)

	core := make(map[int]float64, len(graph))
	removed := make(map[int]bool, len(graph))

	// peeling process
	for hArr.Len() > 0 {
		node := heap.Pop(&hArr).(*Node)
		u, d := node.id, node.deg
		// skip if already removed
		if removed[u] {
			continue
		}
		core[u] = d
		removed[u] = true
		// decrement degree of neighbors
		for _, v := range graph[u] {
			if !removed[v] {
				degree[v]--
				hArr.update(nodes[v], degree[v])
			}
		}
	}

	for i, value := range core {
		fmt.Fprintf(outputFile, "%d: %.4f\n", i, value)
	}
}
