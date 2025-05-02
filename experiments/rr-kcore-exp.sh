#!/bin/sh
#SBATCH --job-name=rr-kcore-bigmem
#SBATCH --partition=bigmem
#SBATCH --time=1-00:00:00
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=20
#SBATCH --mem=0

ml Go

cd ../cmd/
# for graph in 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp' 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
do
     for alg in 'rr-kcore'
    do
        cd ../scripts/
        echo "Partitioining Graph: $graph"
        python3 graph_partitioner.py $graph 80 /home/pm886/palmer_scratch/graph-dp-experiments/graphs_new/
        cd ../cmd/
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
        cd ../experiments/
    done
    cd ../scripts/
    echo "Cleanup Graph: $graph"
    python3 cleanup.py $graph 80 /home/pm886/palmer_scratch/graph-dp-experiments/graphs_new/
    cd ../experiments/
done
