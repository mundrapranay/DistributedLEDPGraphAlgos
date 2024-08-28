package algorithms

import (
	"fmt"
	"math"
	"os"
	"time"

	datastructures "github.com/mundrapranay/DistributedLEDPGraphAlgos/data-structures"
	distribution "github.com/mundrapranay/DistributedLEDPGraphAlgos/noise"
)

func KCoreCDPCoord(n int, phi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, graphFileName string, outputFileName string) {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	startTime := time.Now()
	levels_per_group := math.Ceil(logAToBaseB(n, 1.0+phi)) / 4
	rounds_param := math.Ceil(4.0 * math.Pow(logAToBaseB(n, 1.0+phi), 1.2))
	number_of_rounds := int(rounds_param)
	lambda := 0.5
	super_step1_geom_factor := epsilon * factor
	super_step2_geom_factor := epsilon * (1.0 - factor)

	graph := loadGraphWorker(graphFileName, 0, super_step1_geom_factor, levels_per_group, bias, bias_factor, noise, true)
	lds := datastructures.NewLDS(n, levels_per_group)
	preProcessingTime := time.Now()
	preTime := preProcessingTime.Sub(startTime)
	fmt.Fprintf(outputFile, "Preprocessing Time: %.8f\n", preTime.Seconds())

	for round := 0; round < number_of_rounds-2; round++ {
		group_index := lds.GroupForLevel(uint(round))

		for _, vertex := range graph {
			if vertex.round_threshold == round {
				vertex.permanent_zero = 0
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
					scale := super_step2_geom_factor / (2.0 * float64(vertex.round_threshold))
					geomDist := distribution.NewGeomDistribution(scale)
					noise_sampled := geomDist.TwoSidedGeometric()
					extra_bias := int64(3 * (2 * math.Exp(scale)) / math.Pow((math.Exp(2*scale)-1), 3))
					noised_neighbor_count += noise_sampled
					noised_neighbor_count += extra_bias
				}

				if noised_neighbor_count > int64(math.Pow((1+phi), float64(group_index))) {
					vertex.next_level = 1
				} else {
					vertex.permanent_zero = 0
				}
			}
		}

		for _, vertex := range graph {
			if vertex.next_level == 1 && vertex.permanent_zero != 0 {
				lds.LevelIncrease(uint(vertex.id))
			}
		}
	}

	estimated_core_numbers := estimateCoreNumbers(lds, n, phi, lambda, float64(levels_per_group))
	endTime := time.Now()
	for i, value := range estimated_core_numbers {
		fmt.Fprintf(outputFile, "%d: %.4f\n", i, value)
	}
	algoTime := endTime.Sub(preProcessingTime)
	fmt.Fprintf(outputFile, "Algorithm Time: %.8f\n", algoTime.Seconds())
	outputFile.Close()
}

func KCoreCDPACount(n int, phi float64, epsilon float64, factor float64, bias bool, bias_factor int, noise bool, graphFileName string) *datastructures.LDS {
	levels_per_group := math.Ceil(logAToBaseB(n, 1.0+phi)) / 4
	rounds_param := math.Ceil(4.0 * math.Pow(logAToBaseB(n, 1.0+phi), 1.2))
	number_of_rounds := int(rounds_param)
	super_step1_geom_factor := epsilon * factor
	super_step2_geom_factor := epsilon * (1.0 - factor)

	graph := loadGraphWorker(graphFileName, 0, super_step1_geom_factor, levels_per_group, bias, bias_factor, noise, true)
	lds := datastructures.NewLDS(n, levels_per_group)
	for round := 0; round < number_of_rounds-2; round++ {
		group_index := lds.GroupForLevel(uint(round))

		for _, vertex := range graph {
			if vertex.round_threshold == round {
				vertex.permanent_zero = 0
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
					// geomDist := datastructures.NewGeomDistribution(lambda/(2.0*rounds_param), 1.0)
					// noised_neighbor_count += int(geomDist.DoubleGeomSample())
					scale := super_step2_geom_factor / (2.0 * float64(vertex.round_threshold))
					// scale := 0.5
					geomDist := distribution.NewGeomDistribution(scale)
					noise_sampled := geomDist.TwoSidedGeometric()
					// fmt.Printf("%.4f\n", scale)
					// noise_sampled2, err := googlenoise.Laplace().AddNoiseInt64(noised_neighbor_count, 1, 1, scale, 0)
					// if err != nil {
					// 	fmt.Println(err.Error())
					// 	// return
					// }
					extra_bias := int64(3 * (2 * math.Exp(scale)) / math.Pow((math.Exp(2*scale)-1), 3))
					// fmt.Printf("Worker: %d | Round: %d | Vertex: %d | Noise Sampled: %d\n", workerID, round, vertex.id, noise_sampled)
					noised_neighbor_count += noise_sampled
					noised_neighbor_count += extra_bias
				}

				if noised_neighbor_count > int64(math.Pow((1+phi), float64(group_index))) {
					vertex.next_level = 1
				} else {
					vertex.permanent_zero = 0
				}
			}
		}

		for _, vertex := range graph {
			if vertex.next_level == 1 && vertex.permanent_zero != 0 {
				lds.LevelIncrease(uint(vertex.id))
			}
		}
	}

	return lds
}
