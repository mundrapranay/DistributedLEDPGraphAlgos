##!/bin/sh
#cd
#mkdir results
#mkdir graph-dp-experiments
#cd graph-dp-experiments
#mkdir graphs
#cd
#wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
#sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
#export PATH=$PATH:/usr/local/go/bin
#go version
#cd localgraph-dp/
#go mod tidy
#cd


#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Create directories for results and graphs
mkdir -p results
mkdir -p graph-dp-experiments/graphs

# Assume experiments/config/ already exists.

# Check if the --download-big flag is provided to download large graphs
DOWNLOAD_BIG=false
if [ "$1" = "--download-big" ]; then
    DOWNLOAD_BIG=true
fi

# Change to the graphs directory and download the small graphs
cd ./graph-dp-experiments/graphs
echo "Downloading small graphs..."
wget -c https://storage.googleapis.com/ledp-graphs/graphs/email-eu-core_adj_f -O email-eu-core_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/wiki_adj_f -O wiki_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/enron_adj_f -O enron_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/brightkite_adj_f -O brightkite_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/ego-twitter_adj_f -O ego-twitter_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/gplus_adj_f -O gplus_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/stanford_adj_f -O stanford_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/dblp_adj -O dblp_adj
#wget -c https://storage.googleapis.com/ledp-graphs/graphs/brain_adj -O brain_adj

# Optionally download the large graphs (the last 4)
if [ "$DOWNLOAD_BIG" = true ]; then
    echo "Downloading large graphs..."
    wget -c https://storage.googleapis.com/ledp-graphs/graphs/orkut_adj -O orkut_adj
    wget -c https://storage.googleapis.com/ledp-graphs/graphs/livejournal_adj -O livejournal_adj
    wget -c https://storage.googleapis.com/ledp-graphs/graphs/twitter_adj -O twitter_adj
    wget -c https://storage.googleapis.com/ledp-graphs/graphs/friendster_adj -O friendster_adj
fi
cd ../../

# Detect system architecture for Go installation
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    GO_ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    GO_ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

GO_VERSION="1.22.2"
GO_TAR="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
echo "Downloading Go ${GO_VERSION} for ${GO_ARCH}..."
wget -c https://go.dev/dl/${GO_TAR}
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ${GO_TAR}
export PATH=$PATH:/usr/local/go/bin
go version

# Generate YAML configuration files for experiments
echo "Generating configuration files..."
declare -A GRAPH_SIZES=(
    ["email-eu-core"]=986
    ["wiki"]=7115
    ["enron"]=36692
    ["brightkite"]=58228
    ["ego-twitter"]=81306
    ["gplus"]=107614
    ["stanford"]=281903
    ["dblp"]=317080
    ["brain"]=784262
    ["orkut"]=3072441
    ["livejournal"]=4846609
    ["twitter"]=41652230
    ["friendster"]=65608366
)
ALGO_NAMES=("kcoreLDP" "triangle_countingLDP")

for graph in "${!GRAPH_SIZES[@]}"; do
    for algo in "${ALGO_NAMES[@]}"; do
        CONFIG_FILE="./experiments/config/${graph}-${algo}.yaml"
        cat > "$CONFIG_FILE" <<EOF
graph: ${graph}
graph_size: ${GRAPH_SIZES[$graph]}
algo_name: ${algo}
num_workers: 80
epsilon: 0.5
phi: 0.5
runs: 1
bias: true
bias_factor: 8
noise: true
output_file_tag: hpc_baseline_true
graph_loc: ../graph-dp-experiments/graphs/
EOF
        echo "Created config file: $CONFIG_FILE"
    done
done

echo "Setup complete."

