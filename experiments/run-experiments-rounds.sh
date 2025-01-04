#!/bin/sh
#SBATCH --job-name=kcoreLDP
#SBATCH --partition=day
#SBATCH --time=1-00:00:00
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=48
#SBATCH --mem=0

ml Go

cd ../cmd/
# for graph in 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp' 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
do
    # for alg in 'kcoreLDP' 'triangle_countingLDP'
    for alg in 'kcoreLDP'
    do
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
    done
done