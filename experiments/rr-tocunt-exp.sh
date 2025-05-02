#!/bin/sh
#SBATCH --job-name=rr-tcount-bigmem
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
for graph in 'email-eu-core' 'wiki' 'enron' 'brightkite' 'ego-twitter' 'gplus' 'stanford' 'dblp'
do
     for alg in 'rr-tcount'
    do
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
    done
done
