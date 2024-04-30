package algorithms

import (
	"fmt"
	"math"
	go_rand "math/rand"
	"os"
	"sort"
	"time"

	google_laplace "github.com/google/differential-privacy/go/v2/noise"
	distribution "github.com/mundrapranay/DistributedLEDPGraphAlgos/noise"
)

func TriangleCountingCDP(n int, psi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, graphFileName string, outputFileName string) {

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	startTime := time.Now()
	lds := KCoreCDPACount(n, psi, epsilon/4, factor, bias, bias_factor, noise, graphFileName)
	kcoreTime := time.Now()
	kcore_time := kcoreTime.Sub(startTime)
	fmt.Fprintf(outputFile, "KCore Time: %.8f\n", kcore_time.Seconds())

	X := make([][]bool, n)
	for i := range X {
		X[i] = make([]bool, n)
	}

	graph := loadGraphTCount(graphFileName, 0, true)
	preProcessingTime := time.Now()
	preTime := preProcessingTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Preprocessing Time: %.8f\n", preTime.Seconds())

	// main logic publish RR
	// parallelize
	// noised_neighbours := make([][]int, n)
	for id, neighbours := range graph {
		var noised_neighbours []int
		if noise {
			neighbours_rr := randomizedResponse(epsilon/4, neighbours, n, id)
			sort.Ints(neighbours_rr)
			noised_neighbours = neighbours_rr
		} else {
			sort.Ints(neighbours)
			noised_neighbours = neighbours
		}

		for _, j := range noised_neighbours {
			X[id][j] = true
			X[j][id] = true
		}
	}

	// for CDP-Parallel

	// var wg sync.WaitGroup
	// // var mu sync.Mutex
	// // numGoroutines := runtime.NumCPU() // use the number of cores available for the process
	// numGoroutines := runtime.NumCPU()
	// // Calculate the workload for each goroutine
	// perGoroutine := n / numGoroutines

	// for g := 0; g < numGoroutines; g++ {
	// 	start := g * perGoroutine
	// 	end := start + perGoroutine
	// 	if g == numGoroutines-1 {
	// 		end = n // Ensure the last goroutine covers the remainder
	// 	}

	// 	wg.Add(1)
	// 	go func(start, end int) {
	// 		defer wg.Done()
	// 		// localDebug := 0

	// 		// i runs from [0, end)
	// 		// j runs from [i + 1, n-1)
	// 		// k runs from [j + 1, n)
	// 		// these loop boundaries avoid recomputation as we have an undirected graph
	// 		// edgeCount := 0
	// 		for i := start; i < end; i++ {
	// 			neighbours := graph[i]
	// 			if noise {
	// 				neighbours_rr := randomizedResponse_mem_optim(epsilon/4, neighbours, n, i)
	// 				sort.Ints(neighbours_rr)
	// 				noised_neighbours[i] = neighbours_rr
	// 			} else {
	// 				sort.Ints(neighbours)
	// 				noised_neighbours[i] = neighbours
	// 			}
	// 		}
	// 	}(start, end)
	// }

	// wg.Wait()

	// for i, nghs := range noised_neighbours {
	// 	for _, j := range nghs {
	// 		X[i][j] = true
	// 		X[j][i] = true
	// 	}
	// }

	superStepTime := time.Now()
	fmt.Fprintf(outputFile, "Publish RR Time: %.8f\n", superStepTime.Sub(preProcessingTime).Seconds())

	OEtime := time.Now()
	max_noisy_out_degree := 0.0
	for id, neighbours := range graph {
		// only keep outgoing edges
		var outgoing_edges []int
		node_level, err := lds.GetLevel(uint(id))
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, neighbour := range neighbours {
			j_level, err := lds.GetLevel(uint(neighbour))
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
		geomDist := distribution.NewGeomDistribution(epsilon / 4)
		noisy_out_degree := outDegree + float64(geomDist.TwoSidedGeometric())
		if noisy_out_degree > max_noisy_out_degree {
			max_noisy_out_degree = noisy_out_degree
		}
	}
	fmt.Fprintf(outputFile, "Max Noisy Outdegree: %.18f\n", max_noisy_out_degree)

	var b2i = map[bool]float64{false: 0, true: 1}
	// compute tcount and publish
	triangle_count := 0.0
	u := math.Exp(epsilon/4) + 1.0
	denom := (math.Exp(epsilon/4) - 1)
	for id, neighbours := range graph {
		localTCount := 0.0
		// only keep outgoing edges
		var outgoing_edges []int
		node_level, err := lds.GetLevel(uint(id))
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, neighbour := range neighbours {
			j_level, err := lds.GetLevel(uint(neighbour))
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
		end := int(math.Min(max_noisy_out_degree, float64(len(outgoing_edges))))
		for j := 0; j < end; j++ {
			for k := j + 1; k < end; k++ {
				localTCount += (b2i[X[outgoing_edges[j]][outgoing_edges[k]]]*u - 1) / denom
				// localTCount_debug += int(b2i(X[edges[j]][edges[k]]))
			}
		}
		epsilon_lap := epsilon / 4
		localNoisyTcount, err := google_laplace.Laplace().AddNoiseFloat64(localTCount, 1, max_noisy_out_degree, epsilon_lap/2, 0)
		if err != nil {
			fmt.Printf("Not able to sample\n")
		}
		triangle_count += localNoisyTcount
	}
	computeTime := time.Now()
	fmt.Fprintf(outputFile, "Compute Time: %.8f\n", computeTime.Sub(OEtime).Seconds())
	endTime := time.Now()
	fmt.Fprintf(outputFile, "Triangle Count Approx: %.8f\n", triangle_count)
	// fmt.Fprintf(outputFile, "Triangle Count (int): %.8f\n", triangle_count[1])
	algoTime := endTime.Sub(preProcessingTime)
	fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
	outputFile.Close()

}
