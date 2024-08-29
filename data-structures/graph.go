package datastructures

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Graph struct {
	GraphSize     int
	AdjacencyList map[int][]int
}

func NewGraph(filename string, bidirectional bool) (*Graph, error) {
	adjacencyList := make(map[int][]int)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s\n", filename)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Fields(line)

		vertex, err := strconv.Atoi(values[0])
		if err != nil {
			return nil, err
		}

		ngh, err := strconv.Atoi(values[1])
		if err != nil {
			return nil, err
		}

		neighbors1 := adjacencyList[vertex]
		if ngh >= 0 {
			neighbors1 = append(neighbors1, ngh)
		}
		adjacencyList[vertex] = neighbors1

		// if bidirectional edges are needed
		if bidirectional {
			neighbors2 := adjacencyList[ngh]
			neighbors2 = append(neighbors2, vertex)
			adjacencyList[ngh] = neighbors2
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	graphSize := len(adjacencyList)

	return &Graph{
		GraphSize:     graphSize,
		AdjacencyList: adjacencyList,
	}, nil
}
