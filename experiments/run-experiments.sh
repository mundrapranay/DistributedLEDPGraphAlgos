#!/bin/sh


for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp'
do
    for alg in 'kcoreLDP' 'triangle_countingLDP'
    do
        cd ../plots/
        echo "Partitioining Graph: $graph"
        python3 graph_partitioner.py $graph 80
        cd ../cmd/
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
        cd ../experiments/
    done
    cd ../plots/
    echo "Cleanup Graph: $graph"
    python3 cleanup.py $graph 80
    cd ../experiments/
done