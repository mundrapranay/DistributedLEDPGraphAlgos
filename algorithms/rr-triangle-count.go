package algorithms

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	go_rand "math/rand"

	"github.com/dgravesa/go-parallel/parallel"
	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
)

type TVertex struct {
	neighbours []int
}

type TCoordinator struct {
	workerChannels map[int]chan [][]int
	lock           sync.Mutex
	wg             sync.WaitGroup
	worker_wg      sync.WaitGroup
}

func (t_coord *TCoordinator) sendData(workerID int, noised_neighbours [][]int) {
	t_coord.lock.Lock()
	defer t_coord.lock.Unlock()

	if ch, ok := t_coord.workerChannels[workerID]; ok {
		ch <- noised_neighbours
	}
}

func (t_coord *TCoordinator) processData(chunk int, n int, epsilon float64) float64 {
	// fmt.Print("Processing Data\n")
	X := make([][]int, n)
	Y := make([][]float64, n)

	for i := range X {
		Y[i] = make([]float64, n)
	}

	for workerID, ch := range t_coord.workerChannels {
		channel_data := <-ch
		t_coord.lock.Lock()
		// process data
		for vertexID, noised_neighours := range channel_data {
			i := vertexID + workerID*chunk
			X[i] = noised_neighours
		}
		t_coord.lock.Unlock()
		// fmt.Printf("Done processing data for Worker %d\n", workerID)
	}

	// reinit workerChannels to free memory
	t_coord.workerChannels = make(map[int]chan [][]int, 1)

	// @ToDo: parallelize to speed up compute
	// for i := 0; i < n-1; i++ {
	// 	for idx, value := range X[i] {
	// 		j := i + 1 + idx
	// 		if value == 1 && j < n {
	// 			Y[i][j] = int((float64(value)*(math.Exp(epsilon)+1.0) - 1.0) / (math.Exp(epsilon) - 1.0))
	// 		}
	// 		// else {
	// 		// 	fmt.Printf("i: %d, j: %d\n", i, j)
	// 		// }
	// 	}
	// }

	parallel.For(n-1, func(i, _ int) {
		for idx, value := range X[i] {
			j := i + 1 + idx
			if value == 1 && j < n {
				Y[i][j] = (float64(value)*(math.Exp(epsilon)+1.0) - 1.0) / (math.Exp(epsilon) - 1.0)
			}
		}
	})

	// fmt.Printf("Done with computing Y\n")
	// @ToDo: Make it parallel? Lock triangle_count or an array and then sum
	triangle_count_estimate := 0.0
	triangle_count_store := make([]float64, n-2)
	// for i := 0; i < n-2; i++ {
	// 	for j := i + 1; j < n-1; j++ {
	// 		for k := j + 1; k < n; k++ {
	// 			// @ToDo: add additional checks to speed up compute
	// 			if Y[i][j] == 1 {
	// 				triangle_count_estimate += (Y[i][j] * Y[j][k] * Y[i][k])
	// 			}
	// 		}
	// 	}
	// }

	// i,j,k TCount = ((X[i][j] * X[j][k] * X[i][k]) * (math.Exp(epsilon)+1.0) - 1.0) / (math.Exp(epsilon) - 1.0))^3
	// X is hashMap where i,j is key and it exists if i--j is an edge
	// @ToDo: change [(math.Exp(epsilon)+1.0) - 1.0) / (math.Exp(epsilon) - 1.0)] to flaot64 when running
	parallel.For(n-2, func(i, _ int) {
		for j := i + 1; j < n-1; j++ {
			for k := j + 1; k < n; k++ {
				// @ToDo: add additional checks to speed up compute
				// if Y[i][j] == 1 {
				// 	triangle_count_store[i] += (Y[i][j] * Y[j][k] * Y[i][k])
				// }
				triangle_count_store[i] += (Y[i][j] * Y[j][k] * Y[i][k])
			}
		}
	})

	for _, count := range triangle_count_store {
		triangle_count_estimate += count
	}

	return triangle_count_estimate

}

// @ToDo: Fails when #edges ~ 5.7 million
// loadGraph loads the graph from a file.
func t_loadGraph(filename string, offset int, n int) map[int]*TVertex {

	processed_graph := make(map[int]*TVertex)
	graph, err := datastructures.NewGraph(filename, false)
	if err != nil {
		fmt.Printf(err.Error())
	}
	// store only upper triangle [id+1:]
	for node, neighbours := range graph.AdjacencyList {
		neighbours_updated := make([]int, n)
		for _, n_i := range neighbours {
			neighbours_updated[n_i] = 1
		}
		vertex := &TVertex{
			// id:         node,
			neighbours: neighbours_updated[node+1:],
		}
		processed_graph[node-offset] = vertex
	}
	return processed_graph
}

// @ToDo: rand.Uniform() is bottleneck as it's SecureURG
func randomizedResponseRR(epsilon float64, neighbours []int) {
	prob := 1.0 / (math.Exp(epsilon) + 1.0)
	for n := range neighbours {
		if go_rand.Float64() < prob {
			neighbours[n] = 1 - neighbours[n]
		}
	}
}

func t_superStep(workerID int, n int, epsilon float64, offset int, noise bool, graph map[int]*TVertex, t_coordinator *TCoordinator) {
	// @ToDo: check if we are also adding edges, by printing the degree information before and after RR
	noised_neighbours := make([][]int, len(graph))
	for id, t_vertex := range graph {
		if noise {
			randomizedResponseRR(epsilon, t_vertex.neighbours)
		}
		// @ToDo: update the hash map after RR
		noised_neighbours[id] = t_vertex.neighbours
	}

	// @ToDo: only send the IDs of edges that exist after RR
	t_coordinator.sendData(workerID, noised_neighbours)
	t_coordinator.worker_wg.Done()
	fmt.Printf("Worker %d Done\n", workerID)
}

func TriangleCountingRR(n int, epsilon float64, noise bool, baseFileName string, workerFileNames []string, outputFileName string) {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	startTime := time.Now()

	number_of_workers := len(workerFileNames)
	chunk := n / number_of_workers

	t_coordinator := &TCoordinator{
		workerChannels: make(map[int]chan [][]int, number_of_workers),
	}
	t_coordinator.wg.Add(1)

	// two loops can be made into one
	// t_coordinator.worker_wg.Add(number_of_workers)
	var worker_graphs []map[int]*TVertex
	for i := 0; i < number_of_workers; i++ {
		filename := baseFileName + workerFileNames[i]
		offset := i * chunk
		graph := t_loadGraph(filename, offset, n)
		worker_graphs = append(worker_graphs, graph)
		t_coordinator.workerChannels[i] = make(chan [][]int, len(graph))
	}
	// t_coordinator.worker_wg.Wait()
	preProcessingTime := time.Now()
	preTime := preProcessingTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Preprocessing Time: %.8f\n", preTime.Seconds())

	// main logic
	t_coordinator.worker_wg.Add(number_of_workers)
	for i := 0; i < number_of_workers; i++ {
		graph := worker_graphs[i]
		go func(workerID int, graph map[int]*TVertex) {
			offset := workerID * chunk
			t_superStep(workerID, n, epsilon, offset, noise, graph, t_coordinator)
		}(i, graph)
		worker_graphs[i] = make(map[int]*TVertex, 1)
	}

	t_coordinator.worker_wg.Wait()
	superStepTime := time.Now()
	fmt.Fprintf(outputFile, "Superstep Time: %.8f\n", superStepTime.Sub(preProcessingTime).Seconds())

	// @ToDo: delete all the graphs here

	triangle_count := t_coordinator.processData(chunk, n, epsilon)
	computeTime := time.Now()
	fmt.Fprintf(outputFile, "Compute Time: %.8f\n", computeTime.Sub(superStepTime).Seconds())

	for _, ch := range t_coordinator.workerChannels {
		close(ch)
	}
	t_coordinator.wg.Done()
	fmt.Fprintf(outputFile, "Triangle Count Approx: %.8f\n", triangle_count)
	endTime := time.Now()
	algoTime := endTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
}
