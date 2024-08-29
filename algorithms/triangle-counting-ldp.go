package algorithms

import (
	"fmt"
	"math"
	go_rand "math/rand"
	"os"
	"sort"
	"sync"
	"time"

	google_laplace "github.com/google/differential-privacy/go/v2/noise"
	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
	distribution "github.com/mundrapranay/DistributedLEDPGraphAlgos/noise"
	"golang.org/x/exp/slices"
)

type TCountCoordinator struct {
	workerChannelsNeighborsRR map[int]chan [][]int
	lock                      sync.Mutex
	wg                        sync.WaitGroup
	worker_wg                 sync.WaitGroup
	X                         [][]bool
	workerChannelsTCount      map[int]chan float64
	workerChannelsMaxOut      map[int]chan float64
}

func (coord *TCountCoordinator) sendDataRR(workerID int, noised_neighbours [][]int) {
	coord.lock.Lock()
	defer coord.lock.Unlock()

	if ch, ok := coord.workerChannelsNeighborsRR[workerID]; ok {
		ch <- noised_neighbours
	}

}

func (coord *TCountCoordinator) sendDataTCount(workerID int, private_tcount float64) {
	coord.lock.Lock()
	defer coord.lock.Unlock()

	if ch, ok := coord.workerChannelsTCount[workerID]; ok {
		ch <- private_tcount
	}

}

func (coord *TCountCoordinator) sendMaxNoisyOutDegree(workerID int, private_max_dv float64) {
	coord.lock.Lock()
	defer coord.lock.Unlock()

	if ch, ok := coord.workerChannelsMaxOut[workerID]; ok {
		ch <- private_max_dv
	}

}

func (coord *TCountCoordinator) publishNoisyEdges(chunk int) {

	// Assuming workerChannels gets populated somewhere before this
	for workerID, ch := range coord.workerChannelsNeighborsRR {
		channel_data := <-ch
		coord.lock.Lock()
		for vertexID, noised_neighbours := range channel_data {
			i := vertexID + workerID*chunk
			for _, neighbour := range noised_neighbours {
				coord.X[i][neighbour] = true
				coord.X[neighbour][i] = true
			}
		}
		coord.lock.Unlock()
	}
}

func (coord *TCountCoordinator) aggregateCounts() float64 {

	t_count := 0.0
	for _, ch := range coord.workerChannelsTCount {
		channel_data := <-ch
		coord.lock.Lock()
		t_count += channel_data
		coord.lock.Unlock()
	}
	return t_count
}

func (coord *TCountCoordinator) computePublicNoisyOutDegree() float64 {
	// fmt.Print("Counting\n")
	noisy_dv := 0.0
	for _, ch := range coord.workerChannelsMaxOut {
		channel_data := <-ch
		// fmt.Printf("Worker %d %.8f Done\n", id, channel_data)
		coord.lock.Lock()
		if channel_data > noisy_dv {
			noisy_dv = channel_data
		}
		coord.lock.Unlock()
	}
	return noisy_dv
}

func loadGraphTCount(filename string, offset int, bilateral bool) [][]int {
	graph, err := datastructures.NewGraph(filename, bilateral)
	if err != nil {
		fmt.Printf(err.Error())
	}
	processed_graph := make([][]int, graph.GraphSize)
	for node, neighbours := range graph.AdjacencyList {
		processed_graph[node-offset] = neighbours
	}
	return processed_graph
}

func randomizedResponse(epsilon float64, neighbours []int, n int, nodeID int) []int {
	prob := 1.0 / (math.Exp(epsilon) + 1.0)
	var updated_nghs []int

	// // Cryptographically Secure Random Number
	// for i := nodeID + 1; i < n; i++ {
	// 	if rand.Uniform() < prob {
	// 		if !slices.Contains(neighbours, i) {
	// 			updated_nghs = append(updated_nghs, i)
	// 		}
	// 	} else {
	// 		if slices.Contains(neighbours, i) {
	// 			updated_nghs = append(updated_nghs, i)
	// 		}
	// 	}
	// }

	// Pseudo Random Number
	for i := nodeID + 1; i < n; i++ {
		if go_rand.Float64() < prob {
			if !slices.Contains(neighbours, i) {
				updated_nghs = append(updated_nghs, i)
			}
		} else {
			if slices.Contains(neighbours, i) {
				updated_nghs = append(updated_nghs, i)
			}
		}
	}

	return updated_nghs
}

func workerRR(workerID int, n int, epsilon float64, offset int, workLoad int, noise bool, graph [][]int, coordinator *TCountCoordinator) {
	noised_neighbours := make([][]int, workLoad)
	for id, neighbours := range graph {
		if noise {
			neighbours_rr := randomizedResponse(epsilon, neighbours, n, id+offset)
			sort.Ints(neighbours_rr)
			noised_neighbours[id] = neighbours_rr
		} else {
			sort.Ints(neighbours)
			noised_neighbours[id] = neighbours
		}
	}

	coordinator.sendDataRR(workerID, noised_neighbours)
	coordinator.worker_wg.Done()
	// fmt.Printf("Worker RR %d Done\n", workerID)
}

func workerMaxOutDegree(workerID int, n int, epsilon float64, offset int, graph [][]int, lds *datastructures.LDS, coordinator *TCountCoordinator) {
	workerNoisyDv := 0.0
	for id, neighbours := range graph {
		// only keep outgoing edges
		var outgoing_edges []int
		node_level, err := lds.GetLevel(id + offset)
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, neighbour := range neighbours {
			j_level, err := lds.GetLevel(neighbour)
			if err != nil {
				fmt.Println(err.Error())
			}
			if j_level > node_level {
				outgoing_edges = append(outgoing_edges, neighbour)
			} else if j_level == node_level {
				if go_rand.Float64() <= 0.5 {
					outgoing_edges = append(outgoing_edges, neighbour)
				}
			}
		}
		outDegree := float64(len(outgoing_edges))
		geomDist := distribution.NewGeomDistribution(epsilon)
		noisy_out_degree := outDegree + float64(geomDist.TwoSidedGeometric())
		if noisy_out_degree > workerNoisyDv {
			workerNoisyDv = noisy_out_degree
		}
	}

	// t_coordinator.sendData_a_count(workerID, noised_neighbours, outgoing_edges)
	coordinator.sendMaxNoisyOutDegree(workerID, workerNoisyDv)
	coordinator.worker_wg.Done()
	fmt.Printf("Worker Count %d Done\n", workerID)
}

func workerCountTriangles(workerID int, epsilon float64, noisy_out_degree float64, offset int, graph [][]int, lds *datastructures.LDS, coordinator *TCountCoordinator) {
	var b2i = map[bool]float64{false: 0, true: 1}
	workerTCount := 0.0
	u := math.Exp(epsilon) + 1.0
	denom := (math.Exp(epsilon) - 1)
	for id, neighbours := range graph {
		localTCount := 0.0
		// only keep outgoing edges
		var outgoing_edges []int
		node_level, err := lds.GetLevel(id + offset)
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, neighbour := range neighbours {
			j_level, err := lds.GetLevel(neighbour)
			if err != nil {
				fmt.Println(err.Error())
			}
			if j_level > node_level {
				outgoing_edges = append(outgoing_edges, neighbour)
			} else if j_level == node_level {
				if go_rand.Float64() <= 0.5 {
					outgoing_edges = append(outgoing_edges, neighbour)
				}
			}
		}
		sort.Ints(outgoing_edges)
		end := int(math.Min(noisy_out_degree, float64(len(outgoing_edges))))
		for j := 0; j < end; j++ {
			for k := j + 1; k < end; k++ {
				localTCount += (b2i[coordinator.X[outgoing_edges[j]][outgoing_edges[k]]]*u - 1) / denom
			}
		}
		localNoisyTcount, err := google_laplace.Laplace().AddNoiseFloat64(localTCount, 1, noisy_out_degree, epsilon/2, 0)
		if err != nil {
			fmt.Printf("Not able to sample\n")
		}
		workerTCount += localNoisyTcount
	}
	coordinator.sendDataTCount(workerID, workerTCount)
	coordinator.worker_wg.Done()
	// fmt.Printf("Worker Count %d Done\n", workerID)
}

func TCountCoord(n int, phi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, baseFileName string, workerFileNames []string, outputFileName string) {

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	startTime := time.Now()
	//lds := KCoreLDPTCount(n, phi, epsilon/4, factor, bias, bias_factor, noise, baseFileName, workerFileNames)
	var lds datastructures.LDS
	kcoreTime := time.Now()
	kcore_time := kcoreTime.Sub(startTime)
	fmt.Fprintf(outputFile, "KCore Time: %.8f\n", kcore_time.Seconds())

	number_of_workers := len(workerFileNames)
	chunk := n / number_of_workers
	extra := int(n % number_of_workers)

	t_coordinator := &TCountCoordinator{
		workerChannelsNeighborsRR: make(map[int]chan [][]int, number_of_workers),
		workerChannelsTCount:      make(map[int]chan float64, number_of_workers),
		workerChannelsMaxOut:      make(map[int]chan float64, number_of_workers),
		X:                         make([][]bool, n),
	}
	for i := range t_coordinator.X {
		t_coordinator.X[i] = make([]bool, n)
	}
	t_coordinator.wg.Add(1)

	var worker_graphs_v2 [][][]int
	for i := 0; i < number_of_workers; i++ {
		filename := baseFileName + workerFileNames[i]
		offset := i * chunk
		graph_v2 := loadGraphTCount(filename, offset, false)
		worker_graphs_v2 = append(worker_graphs_v2, graph_v2)
		t_coordinator.workerChannelsNeighborsRR[i] = make(chan [][]int, len(graph_v2))
		t_coordinator.workerChannelsTCount[i] = make(chan float64, 1)
		t_coordinator.workerChannelsMaxOut[i] = make(chan float64, 1)
	}
	preProcessingTime := time.Now()
	preTime := preProcessingTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Preprocessing Time: %.8f\n", preTime.Seconds())

	// main logic publish RR
	t_coordinator.worker_wg.Add(number_of_workers)
	for i := 0; i < number_of_workers; i++ {
		graph_v2 := worker_graphs_v2[i]
		var workLoad int
		if i == number_of_workers-1 {
			workLoad = chunk + extra
		} else {
			workLoad = chunk
		}
		go func(workerID int, graph [][]int) {
			offset := workerID * chunk
			workerRR(workerID, n, epsilon/4, offset, workLoad, noise, graph_v2, t_coordinator)
		}(i, graph_v2)
	}

	t_coordinator.worker_wg.Wait()
	t_coordinator.publishNoisyEdges(chunk)
	superStepTime := time.Now()
	fmt.Fprintf(outputFile, "Publish RR Time: %.8f\n", superStepTime.Sub(preProcessingTime).Seconds())

	OEtime := time.Now()

	t_coordinator.worker_wg.Add(number_of_workers)
	for i := 0; i < number_of_workers; i++ {
		graph_v2 := worker_graphs_v2[i]
		go func(workerID int, graph [][]int) {
			offset := workerID * chunk
			workerMaxOutDegree(workerID, n, epsilon/4, offset, graph_v2, &lds, t_coordinator)
		}(i, graph_v2)
		// worker_graphs_v2[i] = make([][]int, 1)
	}
	t_coordinator.wg.Wait()

	max_noisy_out_degree := t_coordinator.computePublicNoisyOutDegree()
	// compute tcount and publish
	t_coordinator.worker_wg.Add(number_of_workers)
	for i := 0; i < number_of_workers; i++ {
		graph_v2 := worker_graphs_v2[i]
		go func(workerID int, graph [][]int) {
			offset := workerID * chunk
			workerCountTriangles(workerID, epsilon/4, max_noisy_out_degree, offset, graph_v2, &lds, t_coordinator)
		}(i, graph_v2)
		worker_graphs_v2[i] = make([][]int, 1)
	}

	t_coordinator.worker_wg.Wait()

	triangle_count := t_coordinator.aggregateCounts()
	computeTime := time.Now()
	fmt.Fprintf(outputFile, "Compute Time: %.8f\n", computeTime.Sub(OEtime).Seconds())
	for _, ch := range t_coordinator.workerChannelsTCount {
		close(ch)
	}
	for _, ch := range t_coordinator.workerChannelsNeighborsRR {
		close(ch)
	}
	t_coordinator.wg.Done()
	endTime := time.Now()
	fmt.Fprintf(outputFile, "Triangle Count Approx: %.8f\n", triangle_count)
	algoTime := endTime.Sub(preProcessingTime)
	fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
	outputFile.Close()

}
