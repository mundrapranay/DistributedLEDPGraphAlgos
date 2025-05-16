#!/bin/sh
#SBATCH --job-name=rr-tcount-bigmem
#SBATCH --partition=bigmem
#SBATCH --time=1-00:00:00
#SBATCH --nodes=2
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=20
#SBATCH --mem=0

ml Go

cd ../cmd/
# for graph in 'brain' 'orkut' 'livejournal' 'twitter' 'friendster'
for graph in  'stanford' 'dblp' 'gplus'
do
     for alg in 'rr-tcount'
    do
        echo "Running Experiments: $graph"
        go run main.go --config_file ${graph}-${alg}.yaml --workers 80
    done
done
