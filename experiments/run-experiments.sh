#!/bin/sh

for num_workers in 21 41 61 81 
do
    for graph in 'dblp' 'gplus' 'wiki' 'brightkite' 'enron' 'brain' 'livejournal' 'orkut' 'twitter' 'friendster'
    do  
        cd ../plots/ 
        echo "Partitioining Graph: $graph"
        python3 graph_partitioner.py $graph $num_workers 
        cd ../cmd/
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}.yaml --workers $num_workers
        cd ../plots/
        echo "Cleanup Graph: $graph"
        python3 cleanup.py $graph $num_workers
        cd ../experiments/
    done
done