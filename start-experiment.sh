#!/bin/sh
#SBATCH --job-name=kcore_mpi
#SBATCH --partition=mpi
#SBATCH --time=8:00:00
#SBATCH --nodes=21
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=8
#SBATCH --mem-per-cpu=8G

sh load-modules.sh
cd cmd
go build -o main main.go
mpirun -np $SLURM_NTASKS ./main