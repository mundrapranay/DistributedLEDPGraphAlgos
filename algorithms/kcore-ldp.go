package algorithms

import (
	"fmt"
	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
	distribution "github.com/mundrapranay/DistributedLEDPGraphAlgos/noise"
	gompi "github.com/sbromberger/gompi"
	"log"
	"math"
	"os"
	"time"
)

type KCoreVertex struct {
	id              int
	current_level   int
	next_level      int
	permanent_zero  int
	round_threshold int
	neighbours      []int
}

//func (coord *KCoreCoordinator) processData(chunk int) {
//	for workerID, ch := range coord.workerChannels {
//
//		// Receive nextLevels from the worker's channel
//		channel_data := <-ch
//		nextLevels := channel_data[0]
//		permanentZero := channel_data[1]
//		coord.lock.Lock()
//		for vertexID, nextLevel := range nextLevels {
//			if nextLevel == 1 && permanentZero[vertexID] != 0 {
//				coord.lds.LevelIncrease(uint(vertexID + workerID*chunk))
//			}
//		}
//		coord.lock.Unlock()
//	}
//}

func updateLevels(workerID int, nextLevels []int32, permanentZeros []int32, chunk int, lds *datastructures.LDS) {
	for vertexID, nextLevel := range nextLevels {
		if nextLevel == 1 && permanentZeros[vertexID] != 0 {
			//log.Printf("Level Increased for Vertex: %v", vertexID+workerID*chunk)
			err := lds.LevelIncrease(uint(vertexID + workerID*chunk))
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
	}
}

// loadGraph loads the graph from a file.
func loadGraphWorker(filename string, offset int, lambda float64, levelsPerGroup float64, bias bool, biasFactor int, noise bool, bidirectional bool) map[int]*KCoreVertex {

	processedGraph := make(map[int]*KCoreVertex)
	graph, err := datastructures.NewGraph(filename, bidirectional)
	if err != nil {
		fmt.Printf(err.Error())
	}

	for node, neighbours := range graph.AdjacencyList {
		degree := len(neighbours)
		noisedDegree := int64(degree)
		if noise {
			geomDist := distribution.NewGeomDistribution(lambda / 2.0)
			noiseSampled := geomDist.TwoSidedGeometric()
			noisedDegree += noiseSampled
			noisedDegree -= int64(math.Min(float64(biasFactor)*float64((2*math.Exp(lambda))/(math.Exp(2*lambda)-1)), float64(noisedDegree)))
			// to ensure degree is atleast 2
			noisedDegree += 1
		}

		threshold := math.Ceil(logAToBaseB(int(noisedDegree), 2)) * levelsPerGroup
		vertex := &KCoreVertex{
			id:              node,
			current_level:   0,
			next_level:      0,
			permanent_zero:  1,
			round_threshold: int(threshold) + 1,
			neighbours:      neighbours,
		}
		processedGraph[node-offset] = vertex
	}
	return processedGraph
}

func workerKCore(workerID int, round int, lambda float64, psi float64, group_index float64, offset int, workLoad int, rounds_param float64, noise bool, graph map[int]*KCoreVertex, currentLevels []int32) ([]int32, []int32) {

	// perform computation for each vertex
	nextLevels := make([]int32, workLoad)
	permanentZeros := make([]int32, workLoad)
	for i := 0; i < len(permanentZeros); i++ {
		permanentZeros[i] = 1
		nextLevels[i] = 0
	}
	for _, vertex := range graph {
		if vertex.round_threshold == round {
			vertex.permanent_zero = 0
			permanentZeros[vertex.id-offset] = 0
		}
		vertexLevel := int(currentLevels[vertex.id])
		vertex.current_level = int(vertexLevel)
		if vertex.current_level == round && vertex.permanent_zero != 0 {
			neighborCount := 0
			for _, ngh := range vertex.neighbours {
				nghLevel := int(currentLevels[ngh])
				if int(nghLevel) == round {
					neighborCount++
				}
			}
			noisedNeighborCount := int64(neighborCount)
			if noise {
				scale := lambda / (2.0 * float64(vertex.round_threshold))
				geomDist := distribution.NewGeomDistribution(scale)
				noiseSampled := geomDist.TwoSidedGeometric()
				extraBias := int64(3 * (2 * math.Exp(scale)) / math.Pow(math.Exp(2*scale)-1, 3))
				noisedNeighborCount += noiseSampled
				noisedNeighborCount += extraBias
			}

			if noisedNeighborCount > int64(math.Pow(1+psi, group_index)) {
				vertex.next_level = 1
				nextLevels[vertex.id-offset] = 1
			} else {
				vertex.permanent_zero = 0
				permanentZeros[vertex.id-offset] = 0
			}
		}
	}
	//data_to_send := [2][]int{nextLevels, permanentZeros}
	//coordinator.sendData(workerID, data_to_send)
	//coordinator.worker_wg.Done()
	return nextLevels, permanentZeros
}

func logAToBaseB(a int, b float64) float64 {
	return math.Log2(float64(a)) / math.Log2(b)
}

func estimateCoreNumbers(lds *datastructures.LDS, n int, phi float64, lambda float64, levels_per_group float64) []float64 {
	coreNumbers := make([]float64, n)
	twoPlusLambda := 2.0 + lambda
	onePlusPhi := 1.0 + phi
	for i := 0; i < n; i++ {
		nodeLevel, err := lds.GetLevel(uint(i))
		if err != nil {
			fmt.Printf(err.Error())
		}
		fracNumerator := nodeLevel + 1.0
		power := math.Max(math.Floor(float64(fracNumerator)/levels_per_group)-1.0, 0.0)
		coreNumbers[i] = twoPlusLambda * math.Pow(onePlusPhi, power)
	}
	return coreNumbers
}

//func KCoreLDPTCount(n int, psi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, baseFileName string, workerFileNames []string) *datastructures.LDS {
//
//	levels_per_group := math.Ceil(log_a_to_base_b(n, 1.0+psi)) / 4
//	rounds_param := math.Ceil(4.0 * math.Pow(log_a_to_base_b(n, 1.0+psi), 1.2))
//	number_of_rounds := int(rounds_param)
//	super_step1_geom_factor := epsilon * factor
//	super_step2_geom_factor := epsilon * (1.0 - factor)
//
//	number_of_workers := len(workerFileNames)
//	chunk := n / number_of_workers
//	extra := n % number_of_workers
//
//	// create a coordinator
//	coordinator := &KCoreCoordinator{
//		lds:            datastructures.NewLDS(n, levels_per_group),
//		workerChannels: make(map[int]chan [2][]int, number_of_workers),
//	}
//	coordinator.wg.Add(1)
//
//	// Preprocess the graphs into an array of workerGraphs.
//	var worker_graphs []map[int]*KCoreVertex
//	for i := 0; i < number_of_workers; i++ {
//		filename := baseFileName + workerFileNames[i]
//		offset := i * chunk
//		graph := loadGraphWorker(filename, offset, super_step1_geom_factor, levels_per_group, bias, bias_factor, noise, false)
//		worker_graphs = append(worker_graphs, graph)
//		coordinator.workerChannels[i] = make(chan [2][]int, len(graph))
//	}
//
//	// main loop
//	for round := 0; round < number_of_rounds-2; round++ {
//
//		group_index := coordinator.lds.GroupForLevel(uint(round))
//		coordinator.worker_wg.Add(number_of_workers)
//
//		for i := 0; i < number_of_workers; i++ {
//
//			graph := worker_graphs[i]
//			go func(workerID int, r int, graph map[int]*KCoreVertex) {
//				// perform computation
//				offset := workerID * chunk
//				var workLoad int
//				if workerID == number_of_workers-1 {
//					workLoad = chunk + extra
//				} else {
//					workLoad = chunk
//				}
//				workerKCore(workerID, r, super_step2_geom_factor, psi, float64(group_index), offset, workLoad, rounds_param, noise, graph, coordinator)
//			}(i, round, graph)
//		}
//
//		// wait for all workers to finish
//		coordinator.worker_wg.Wait()
//
//		// process received messages
//		coordinator.processData(chunk)
//	}
//
//	for _, ch := range coordinator.workerChannels {
//		close(ch)
//	}
//
//	return coordinator.lds
//}

func KCoreLDPCoord(n int, phi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, baseFileName string, workerFileNames []string, outputFileName string) {

	startTime := time.Now()
	gompi.Start(false)
	defer gompi.Stop()

	comm := gompi.NewCommunicator(nil)
	numberOfWorkers := comm.Size() - 1
	rank := comm.Rank()

	if rank == 0 {
		log.Printf("Running with %d workers and 1 coordinator", numberOfWorkers)
	}

	levelsPerGroup := math.Ceil(logAToBaseB(n, 1.0+phi)) / 4
	roundsParam := math.Ceil(4.0 * math.Pow(logAToBaseB(n, 1.0+phi), 1.2))
	numberOfRounds := int(roundsParam)
	lambda := 0.5
	superStep1GeomFactor := epsilon * factor
	superStep2GeomFactor := epsilon * (1.0 - factor)

	//number_of_workers := len(workerFileNames)
	chunk := n / numberOfWorkers
	extra := n % numberOfWorkers

	// create a coordinator maintained lds
	var lds *datastructures.LDS
	if rank == 0 {
		lds = datastructures.NewLDS(n, levelsPerGroup)
	}

	var graph map[int]*KCoreVertex
	if rank != 0 {
		offset := (rank - 1) * chunk
		graph = loadGraphWorker(baseFileName+workerFileNames[rank-1], offset, superStep1GeomFactor, levelsPerGroup, bias, bias_factor, noise, false)
		log.Printf("Graph Loaded %v by worker: %d", baseFileName+workerFileNames[rank-1], rank)
	}

	log.Printf("Starting main loop, worker %d", rank)
	// main loop
	for round := 0; round < numberOfRounds-2; round++ {
		// coordinator gets current levels & group index, and broadcasts the same
		if rank == 0 {
			currentLevels := make([]int32, n)
			groupIndex := 0.0
			for i := 0; i < n; i++ {
				level, err := lds.GetLevel(uint(i))
				if err != nil {
					log.Fatalf(err.Error())
				}
				currentLevels[i] = int32(level)
			}
			groupIndex = float64(lds.GroupForLevel(uint(round)))
			log.Printf("Round %d, Group Index: %.4f", round, groupIndex)
			// broadcast
			for worker := 1; worker <= numberOfWorkers; worker++ {
				comm.SendInt32s(currentLevels, worker, 2)
				comm.SendFloat64(groupIndex, worker, 3)
				log.Printf("Data sent by coordinator for round %d to worker %d", round, worker)
			}
			//comm.BcastInt32s(currentLevels, 0)
			//comm.BcastFloat64s(groupIndexToSend, 0)
			//log.Printf("Data sent by coordinator for round %d", round)

		} else {
			currentLevelsWorkers, _ := comm.RecvInt32s(0, 2)
			groupIndexWorkers, _ := comm.RecvFloat64(0, 3)
			log.Printf("SIze of currentLevels %d recieved by worker %d", len(currentLevelsWorkers), rank)
			offset := (rank - 1) * chunk
			var workLoad int
			if rank == numberOfWorkers {
				workLoad = chunk + extra
			} else {
				workLoad = chunk
			}
			nextLevels, permanentZeros := workerKCore(rank-1, round, superStep2GeomFactor, phi, groupIndexWorkers, offset, workLoad, roundsParam, noise, graph, currentLevelsWorkers)
			comm.SendInt32s(nextLevels, 0, 0)
			comm.SendInt32s(permanentZeros, 0, 1)
			log.Printf("Data sent by worker %d for round %d", rank, round)
		}

		if rank == 0 {
			for worker := 1; worker <= numberOfWorkers; worker++ {
				var receivedNextLevels, receivedPermanentZeros []int32
				//var st1, st2 gompi.Status
				receivedNextLevels, _ = comm.RecvInt32s(worker, 0)
				receivedPermanentZeros, _ = comm.RecvInt32s(worker, 1)
				//log.Printf("next levels from worker %d status: %d", worker, st1.GetError())
				//log.Printf("permZeros from worker %d status: %d", worker, st2.GetError())
				updateLevels(worker-1, receivedNextLevels, receivedPermanentZeros, chunk, lds)
				//log.Printf("Data received by coordinator from worker %d for round %d", worker, round)
			}
			log.Printf("Done with round %d", round)
		}

		comm.Barrier()
	}

	// estimate core numbers function
	//estimatedCoreNumbers := estimateCoreNumbers(coordinator.lds, n, psi, lambda, float64(levelsPerGroup))
	//endTime := time.Now()
	if rank == 0 {
		estimatedCoreNumbers := estimateCoreNumbers(lds, n, phi, lambda, levelsPerGroup)
		endTime := time.Now()
		outputFile, err := os.Create(outputFileName)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer outputFile.Close()

		for i, value := range estimatedCoreNumbers {
			fmt.Fprintf(outputFile, "%d: %.4f\n", i, value)
		}

		algoTime := endTime.Sub(startTime)
		fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
		outputFile.Close()
	}
	//gompi.Stop()
}
