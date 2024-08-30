package algorithms

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
	distribution "github.com/mundrapranay/DistributedLEDPGraphAlgos/noise"
)

type KCoreVertex struct {
	id              int
	current_level   int
	next_level      int
	permanent_zero  int
	round_threshold int
	neighbours      []int
}

type KCoreCoordinator struct {
	lds            *datastructures.LDS
	workerChannels map[int]chan [2][]int
	lock           sync.Mutex
	wg             sync.WaitGroup
	worker_wg      sync.WaitGroup
}

func (coord *KCoreCoordinator) sendData(workerID int, nextLevels [2][]int) {
	coord.lock.Lock()
	defer coord.lock.Unlock()

	if ch, ok := coord.workerChannels[workerID]; ok {
		ch <- nextLevels
	}
}

func (coord *KCoreCoordinator) processData(chunk int) {
	for workerID, ch := range coord.workerChannels {

		// Receive nextLevels from the worker's channel
		channel_data := <-ch
		nextLevels := channel_data[0]
		permanentZero := channel_data[1]
		coord.lock.Lock()
		for vertexID, nextLevel := range nextLevels {
			if nextLevel == 1 && permanentZero[vertexID] != 0 {
				coord.lds.LevelIncrease(uint(vertexID + workerID*chunk))
			}
		}
		coord.lock.Unlock()
	}
}

// loadGraph loads the graph from a file.
func loadGraphWorker(filename string, offset int, lambda float64, levels_per_group float64, bias bool, bias_factor int, noise bool, bidirectional bool) map[int]*KCoreVertex {

	processed_graph := make(map[int]*KCoreVertex)
	graph, err := datastructures.NewGraph(filename, bidirectional)
	if err != nil {
		fmt.Printf(err.Error())
	}

	for node, neighbours := range graph.AdjacencyList {
		degree := len(neighbours)
		noised_degree := int64(degree)
		if noise {
			geomDist := distribution.NewGeomDistribution(lambda / 2.0)
			noise_sampled := geomDist.TwoSidedGeometric()
			noised_degree += noise_sampled
			noised_degree -= int64(math.Min(float64(bias_factor)*float64((2*math.Exp(lambda))/(math.Exp(2*lambda)-1)), float64(noised_degree)))
			// to ensure degree is atleast 2
			noised_degree += 1
		}

		threshold := math.Ceil(log_a_to_base_b(int(noised_degree), 2)) * levels_per_group
		vertex := &KCoreVertex{
			id:              node,
			current_level:   0,
			next_level:      0,
			permanent_zero:  1,
			round_threshold: int(threshold) + 1,
			neighbours:      neighbours,
		}
		processed_graph[node-offset] = vertex
	}
	return processed_graph
}

func workerKCore(workerID int, round int, lambda float64, psi float64, group_index float64, offset int, workLoad int, rounds_param float64, noise bool, graph map[int]*KCoreVertex, coordinator *KCoreCoordinator, lds *datastructures.LDS) {

	// perform computation for each vertex
	nextLevels := make([]int, workLoad)
	permanentZeros := make([]int, workLoad)
	for i := 0; i < len(permanentZeros); i++ {
		permanentZeros[i] = 1
		nextLevels[i] = 0
	}
	for _, vertex := range graph {
		if vertex.round_threshold == round {
			vertex.permanent_zero = 0
			permanentZeros[vertex.id-offset] = 0
		}
		vertex_level, err := lds.GetLevel(uint(vertex.id))
		if err != nil {
			fmt.Printf(err.Error())
		}
		vertex.current_level = int(vertex_level)
		if vertex.current_level == round && vertex.permanent_zero != 0 {
			neighbor_count := 0
			for _, ngh := range vertex.neighbours {
				ngh_level, err := lds.GetLevel(uint(ngh))
				if err != nil {
					fmt.Printf(err.Error())
				}
				if int(ngh_level) == round {
					neighbor_count++
				}
			}
			noised_neighbor_count := int64(neighbor_count)
			if noise {
				scale := lambda / (2.0 * float64(vertex.round_threshold))
				geomDist := distribution.NewGeomDistribution(scale)
				noise_sampled := geomDist.TwoSidedGeometric()
				extra_bias := int64(3 * (2 * math.Exp(scale)) / math.Pow((math.Exp(2*scale)-1), 3))
				noised_neighbor_count += noise_sampled
				noised_neighbor_count += extra_bias
			}

			if noised_neighbor_count > int64(math.Pow((1+psi), group_index)) {
				vertex.next_level = 1
				nextLevels[vertex.id-offset] = 1
			} else {
				vertex.permanent_zero = 0
				permanentZeros[vertex.id-offset] = 0
			}
		}
	}
	data_to_send := [2][]int{nextLevels, permanentZeros}
	coordinator.sendData(workerID, data_to_send)
	coordinator.worker_wg.Done()
}

func log_a_to_base_b(a int, b float64) float64 {
	return math.Log2(float64(a)) / math.Log2(b)
}

func estimateCoreNumbers(lds *datastructures.LDS, n int, phi float64, lambda float64, levels_per_group float64) []float64 {
	core_numbers := make([]float64, n)
	two_plus_lambda := 2.0 + lambda
	one_plus_phi := 1.0 + phi
	for i := 0; i < n; i++ {
		node_level, err := lds.GetLevel(uint(i))
		if err != nil {
			fmt.Printf(err.Error())
		}
		frac_numerator := node_level + 1.0
		power := math.Max(math.Floor(float64(frac_numerator)/levels_per_group)-1.0, 0.0)
		core_numbers[i] = two_plus_lambda * math.Pow(one_plus_phi, power)
	}
	return core_numbers
}

func KCoreLDPTCount(n int, psi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, baseFileName string, workerFileNames []string) *datastructures.LDS {

	levels_per_group := math.Ceil(log_a_to_base_b(n, 1.0+psi)) / 4
	rounds_param := math.Ceil(4.0 * math.Pow(log_a_to_base_b(n, 1.0+psi), 1.2))
	number_of_rounds := int(rounds_param)
	super_step1_geom_factor := epsilon * factor
	super_step2_geom_factor := epsilon * (1.0 - factor)

	number_of_workers := len(workerFileNames)
	chunk := n / number_of_workers
	extra := n % number_of_workers

	// create a coordinator
	coordinator := &KCoreCoordinator{
		lds:            datastructures.NewLDS(n, levels_per_group),
		workerChannels: make(map[int]chan [2][]int, number_of_workers),
	}
	coordinator.wg.Add(1)

	// Preprocess the graphs into an array of workerGraphs.
	var worker_graphs []map[int]*KCoreVertex
	for i := 0; i < number_of_workers; i++ {
		filename := baseFileName + workerFileNames[i]
		offset := i * chunk
		graph := loadGraphWorker(filename, offset, super_step1_geom_factor, levels_per_group, bias, bias_factor, noise, false)
		worker_graphs = append(worker_graphs, graph)
		coordinator.workerChannels[i] = make(chan [2][]int, len(graph))
	}

	// main loop
	for round := 0; round < number_of_rounds-2; round++ {

		group_index := coordinator.lds.GroupForLevel(uint(round))
		coordinator.worker_wg.Add(number_of_workers)

		for i := 0; i < number_of_workers; i++ {

			graph := worker_graphs[i]
			go func(workerID int, r int, graph map[int]*KCoreVertex) {
				// perform computation
				offset := workerID * chunk
				var workLoad int
				if workerID == number_of_workers-1 {
					workLoad = chunk + extra
				} else {
					workLoad = chunk
				}
				workerKCore(workerID, r, super_step2_geom_factor, psi, float64(group_index), offset, workLoad, rounds_param, noise, graph, coordinator, coordinator.lds)
			}(i, round, graph)
		}

		// wait for all workers to finish
		coordinator.worker_wg.Wait()

		// process received messages
		coordinator.processData(chunk)
	}

	for _, ch := range coordinator.workerChannels {
		close(ch)
	}

	return coordinator.lds
}

func KCoreLDPCoord(n int, psi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, baseFileName string, workerFileNames []string, outputFileName string) {

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	startTime := time.Now()
	levels_per_group := math.Ceil(log_a_to_base_b(n, 1.0+psi)) / 4
	rounds_param := math.Ceil(4.0 * math.Pow(log_a_to_base_b(n, 1.0+psi), 1.2))
	number_of_rounds := int(rounds_param)
	lambda := 0.5
	super_step1_geom_factor := epsilon * factor
	super_step2_geom_factor := epsilon * (1.0 - factor)

	number_of_workers := len(workerFileNames)
	chunk := n / number_of_workers
	extra := n % number_of_workers

	// create a coordinator
	coordinator := &KCoreCoordinator{
		lds:            datastructures.NewLDS(n, levels_per_group),
		workerChannels: make(map[int]chan [2][]int, number_of_workers),
	}
	coordinator.wg.Add(1)

	// Preprocess the graphs into an array of workerGraphs.
	var worker_graphs []map[int]*KCoreVertex
	for i := 0; i < number_of_workers; i++ {
		filename := baseFileName + workerFileNames[i]
		offset := i * chunk
		graph := loadGraphWorker(filename, offset, super_step1_geom_factor, levels_per_group, bias, bias_factor, noise, false)
		worker_graphs = append(worker_graphs, graph)
		coordinator.workerChannels[i] = make(chan [2][]int, len(graph))
	}
	preProcessingTime := time.Now()
	preTime := preProcessingTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Preprocessing Time: %.8f\n", preTime.Seconds())

	// main loop
	for round := 0; round < number_of_rounds-2; round++ {

		group_index := coordinator.lds.GroupForLevel(uint(round))
		coordinator.worker_wg.Add(number_of_workers)

		for i := 0; i < number_of_workers; i++ {

			graph := worker_graphs[i]
			go func(workerID int, r int, graph map[int]*KCoreVertex) {
				// perform computation
				offset := workerID * chunk
				var workLoad int
				if workerID == number_of_workers-1 {
					workLoad = chunk + extra
				} else {
					workLoad = chunk
				}
				workerKCore(workerID, r, super_step2_geom_factor, psi, float64(group_index), offset, workLoad, rounds_param, noise, graph, coordinator, coordinator.lds)
			}(i, round, graph)
		}

		// wait for all workers to finish
		coordinator.worker_wg.Wait()

		// process received messages
		coordinator.processData(chunk)
	}

	for _, ch := range coordinator.workerChannels {
		close(ch)
	}

	// estimate core numbers function
	estimated_core_numbers := estimateCoreNumbers(coordinator.lds, n, psi, lambda, float64(levels_per_group))
	endTime := time.Now()
	for i, value := range estimated_core_numbers {
		fmt.Fprintf(outputFile, "%d: %.4f\n", i, value)
	}
	algoTime := endTime.Sub(preProcessingTime)
	fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
	outputFile.Close()
}
