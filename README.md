# DistributedLEDPGraphAlgos

This repository contains a golang implementation for Practical and Accurate Local Differentially Private Graph Algorithms.

## Code Structure

The code base is organized into several directories:

- **algorithms/:** This directory houses all algorithm implementations.

- **cmd/:** The main driver function, `main.go`, is located here.

- **data-structures/:** Implementations of all data structures required by the algorithms can be found in this directory.

- **experiments/:** This directory contains the code and configs for experiments.

## Set-Up

Run `sh setup.sh`

We install golang, and make the necessary data and graph directory. All output files are stored in `${HOME}/results/` and graph files are stored as `${graph-name}_adj` in `${HOME}/graph-dp-experiments/graphs/`.

## Graphs
### k-Core Decomposition LDP
- DBLP : 
- Brain : 
- Orkut : 
- Livejournal :
- Twitter :
- Friendster : 


### Triangle Counting LDP
- Wiki : https://snap.stanford.edu/data/wiki-Vote.html
- Enron : https://snap.stanford.edu/data/email-Enron.html
- Brightkite : https://snap.stanford.edu/data/loc-Brightkite.html
- Gplus : https://snap.stanford.edu/data/ego-Gplus.html
- DBLP :

Note: Use the following script the format Gplus: https://github.com/TriangleLDP/TriangleLDP/blob/main/python/ReadGPlus.py

## Running Experiments

To run an experiment, generate your own YAML config file in the `experiments/configs/` directory . Once you have your config file, go to the `cmd/` folder and run the following command:

```bash
go run main.go -config_file ${name_of_new_config_file} --workers ${number of workers}
```

Note that you only need to provide the name of the config file, not the path.


**Sample YAML Config File (`experiments/configs/twitter.yaml`):**

```yaml
graphs:
  - "twitter"
graph_sizes:
  - 41652230
algo_name: kcoreLDP 
num_workers: 81
eta: 0.9
epsilon: 0.5
psi: 0.5
bias: true
bias_factor: 8
runs: 5
noise: true
output_file_tag: "with_noise_gcp"
graph_loc: "/home/ubuntu/graph-dp-experiments/graphs"
```

`For algo_name, we have the following options: ${kcoreLDP, kcoreCDP, triangle_counting}`

### Reproduce Experimental Results

To reproduce our experimental results, go to `experiments/` run the following command:

```bash
sh run-experiments.sh
```

## Contributing

Feel free to contribute by submitting bug reports, feature requests, or pull requests. We welcome collaboration to enhance the functionality and performance of this implementation.

## License

This project is licensed under the [MIT License](LICENSE).
