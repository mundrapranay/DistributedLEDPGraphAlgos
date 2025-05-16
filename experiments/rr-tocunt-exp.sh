#!/bin/sh
#SBATCH --job-name=rr-tcount-bigmem
#SBATCH --partition=bigmem
#SBATCH --time=1-00:00:00
#SBATCH --nodes=3
#SBATCH --ntasks=3
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=13
#SBATCH --mem=0
#SBATCH --array=0-2

ml Go

cd ../cmd/

# Define the graphs as an array
graphs=('stanford' 'dblp' 'gplus')
alg='rr-tcount'

# Get the current graph based on array task ID
current_graph=${graphs[$SLURM_ARRAY_TASK_ID]}

echo "Running Experiments for graph: $current_graph"
go run main.go --config_file ${current_graph}-${alg}.yaml --workers 80
