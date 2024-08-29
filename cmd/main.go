package main

import (
	"flag"
	"fmt"

	"github.com/mundrapranay/DistributedLEDPGraphAlgos/experiments"
)

func main() {
	var configFile string
	var workers int
	flag.StringVar(&configFile, "config_file", "dblp.yaml", "path to the configuration file")
	flag.IntVar(&workers, "workers", 81, "number of workers")
	flag.Parse()

	if configFile == "" {
		fmt.Println("Usage: go run main.go -config_file <name_of_config_file>")
		return
	}

	experiments.Runner(configFile, workers)

	// @todo: test the following in parallel and see if
	// 		parallel causes random number generator to fail
	// nodeID := 8
	// nghs := []int{9, 10}
	// updated_ngh := experiments.RandomizedResponse(1.0, nghs, n, nodeID)
	// fmt.Printf("%v\n", updated_ngh)
	// n := 17
	// graph := map[int][]int{
	// 	0:  {13, 14},
	// 	1:  {14, 15},
	// 	2:  {3, 15},
	// 	3:  {15},
	// 	4:  {6},
	// 	5:  {6, 16},
	// 	6:  {7, 11},
	// 	7:  {11},
	// 	8:  {9, 10},
	// 	9:  {10},
	// 	10: {15},
	// 	11: {15},
	// 	12: {13, 14},
	// 	13: {14},
	// 	14: {15},
	// 	15: {},
	// 	16: {},
	// }

	// var wg sync.WaitGroup
	// numGoroutines := 17
	// for g := 0; g < numGoroutines; g++ {
	// 	wg.Add(1)
	// 	go func(workerID int) {
	// 		defer wg.Done()
	// 		updated_nghs_local := experiments.RandomizedResponse(1.0, graph[workerID], n, workerID, workerID)
	// 		fmt.Printf("NodeID: %d \t %v\n", workerID, updated_nghs_local)
	// 	}(g)
	// }
	// wg.Wait()

}
