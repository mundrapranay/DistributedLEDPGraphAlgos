#!/bin/sh
#SBATCH --job-name=kcoreLDP
#SBATCH --partition=mpi
#SBATCH --time=00:10:00
#SBATCH --nodes=64
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=4
#SBATCH --mem=0

#sh load-modules.sh
ml Go/1.21.4 OpenMPI/4.1.4-GCC-12.2.0
cd cmd
go build -o main main.go
mpirun -np $SLURM_NTASKS ./main
