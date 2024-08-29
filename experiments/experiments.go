package experiments

import (
	"fmt"
	"io/ioutil"

	"github.com/mundrapranay/DistributedLEDPGraphAlgos/algorithms"

	"gopkg.in/yaml.v2"
)

type ExpirementConfig struct {
	Graphs     []string `yaml:"graphs"`
	GraphSizes []int    `yaml:"graph_sizes"`
	AlgoName   string   `yaml:"algo_name"`
	NumWorkers int      `yaml:"num_workers"`
	Epsilon    float64  `yaml:"epsilon"`
	Phi        float64  `yaml:"phi"`
	Bias       bool     `yaml:"bias"`
	BiasFactor int      `yaml:"bias_factor"`
	Runs       int      `yaml:"runs"`
	Noise      bool     `yaml:"noise"`
	ExpTag     string   `yaml:"output_file_tag"`
	GraphLoc   string   `yaml:"graph_loc"`
}

func Runner(fileName string, workers int) {
	var b2i = map[bool]int8{false: 0, true: 1}
	exp_config := ExpirementConfig{}
	config_file := fmt.Sprintf("../experiments/configs/%s", fileName)
	file, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Printf(err.Error())
	}

	err = yaml.Unmarshal(file, &exp_config)
	if err != nil {
		fmt.Printf("error unmarshalling YAML: %v\n", err)
	}

	exp_config.NumWorkers = workers
	var workerFilesNames []string
	for i := 0; i < exp_config.NumWorkers; i++ {
		workerFilesNames = append(workerFilesNames, fmt.Sprintf("%d.txt", i))
	}

	factor := float64(4.0 / 5.0)
	for run_id := 0; run_id < exp_config.Runs; run_id++ {
		for index, graph := range exp_config.Graphs {
			graph_size := exp_config.GraphSizes[index]
			var outputFile string
			graph_loc := fmt.Sprintf("%s/%s", exp_config.GraphLoc, graph)
			baseFileName := exp_config.GraphLoc
			for bf := exp_config.BiasFactor; bf <= exp_config.BiasFactor; bf++ {
				outputFile = fmt.Sprintf("/home/pm886/palmer_scratch/%s_%s_%.2f_%d_%d_%d_%d_%d_%.2f_%s.txt", graph, exp_config.AlgoName, factor, b2i[exp_config.Bias], b2i[exp_config.Noise], bf, run_id, exp_config.NumWorkers, exp_config.Epsilon, exp_config.ExpTag)
				if exp_config.AlgoName == "kcoreCDP" {
					algorithms.KCoreCDPCoord(graph_size, exp_config.Phi, exp_config.Epsilon, factor, exp_config.Bias, bf, exp_config.Noise, graph_loc, outputFile)
				} else if exp_config.AlgoName == "kcoreLDP" {
					algorithms.KCoreLDPCoord(graph_size, exp_config.Phi, exp_config.Epsilon, factor, exp_config.Bias, bf, exp_config.Noise, baseFileName, workerFilesNames, outputFile)
				} else if exp_config.AlgoName == "triangle_countingLDP" {
					algorithms.TCountCoord(graph_size, exp_config.Phi, exp_config.Epsilon, factor, exp_config.Bias, exp_config.BiasFactor, exp_config.Noise, baseFileName, workerFilesNames, outputFile)
				} else if exp_config.AlgoName == "triangle_countingCDP" {
					algorithms.TriangleCountingCDP(graph_size, exp_config.Phi, exp_config.Epsilon, factor, exp_config.Bias, exp_config.BiasFactor, exp_config.Noise, graph_loc, outputFile)
				}
				fmt.Printf("Done with Exp:%s_%.2f_%t_%d\n", graph, factor, exp_config.Bias, bf)
			}
		}
	}
}
