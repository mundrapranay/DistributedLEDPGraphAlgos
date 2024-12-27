#!/bin/sh

for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp' 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
do
    for alg in 'kcoreLDP'
    do
        if [ "$alg" = "triangle_countingLDP" ] && [ "$graph" != "email-eu-core" ] && [ "$graph" != "wiki" ] && [ "$graph" != "enron" ] && [ "$graph" != "brightkite" ] && [ "$graph" != "ego-twitter" ] && [ "$graph" != "gplus" ] && [ "$graph" != "stanford" ] && [ "$graph" != "dblp" ]; then
            # Skip triangle counting for graphs after dblp
            continue
        fi
        # echo $graph $alg
        cd ../scripts/
        echo "Partitioning Graph: $graph"
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