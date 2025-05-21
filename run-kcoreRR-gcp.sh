#! /bin/bash
cd ../cmd/

for graph in 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
do
     for alg in 'rr-kcore'
    do
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
    done
done