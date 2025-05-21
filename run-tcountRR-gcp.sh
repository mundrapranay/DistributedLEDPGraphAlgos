#! /bin/bash
cd ../cmd/

for graph in 'gplus' 'ego-twitter' 'stanford' 'dblp'
do
     for alg in 'rr-tcount'
    do
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
    done
done