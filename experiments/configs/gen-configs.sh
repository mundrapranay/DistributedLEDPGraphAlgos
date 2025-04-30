#!/bin/bash

# Array of graph names
graphs=("email-eu-core" "wiki" "enron" "brightkite" "ego-twitter" "gplus" "stanford" "dblp" "brain" "orkut" "livejournal" "twitter" "friendster")
graph_sizes=(986 7115 36692 58228 81306 107614 281903 317080 784262 3072441 4846609 41652230 65608366)


# Loop through each graph
for index in "${!graphs[@]}"; do
    graph=${graphs[$index]}
    # Create N files for each graph
    for alg in 'rr-tcount'; do
        filename="${graph}-${alg}.yaml"
        echo "graph: ${graph}" > "$filename"
        echo "graph_size: ${graph_sizes[$index]}" >> "$filename"
        echo "algo_name: ${alg}" >> "$filename"
        echo "num_workers: 80" >> "$filename"
        echo "epsilon: 1.0" >> "$filename"
        echo "phi: 0.5" >> "$filename"
        echo "runs: 1" >> "$filename"
        echo "bias: true" >> "$filename"
        echo "bias_factor: 8" >> "$filename"
        echo "noise: true" >> "$filename"
        echo "output_file_tag: rr_baseline" >> "$filename"
        echo "graph_loc: /home/pm886/palmer_scratch/graph-dp-experiments/graphs_new" >> "$filename"
        echo "Created $filename"
    done
done

echo "All YAML files have been created."

