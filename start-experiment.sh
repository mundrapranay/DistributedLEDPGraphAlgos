#!/bin/sh
#SBATCH --job-name=kcoreLDP
#SBATCH --partition=day
#SBATCH --time=8:00:00
#SBATCH --nodes=21
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=8
#SBATCH --mem-per-cpu=16G

#sh load-modules.sh
ml Go/1.21.4 OpenMPI/4.1.4-GCC-12.2.0
cd cmd
go build -o main main.go
mpirun -np $SLURM_NTASKS ./main