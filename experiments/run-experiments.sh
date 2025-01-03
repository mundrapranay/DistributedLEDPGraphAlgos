#!/bin/sh

# for graph in 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp' 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
do
    # for alg in 'kcoreLDP' 'triangle_countingLDP'
    for alg in 'kcoreLDP'
    do
        cd ../scripts/
        echo "Partitioining Graph: $graph"
        python3 graph_partitioner.py $graph 80
        cd ../cmd/
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
        cd ../experiments/
    done
    cd ../scripts/
    echo "Cleanup Graph: $graph"
    python3 cleanup.py $graph 80
    cd ../experiments/
done