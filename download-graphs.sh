#! /bin/bash

# Download the graphs

cd graphs_new  
wget -c https://storage.googleapis.com/ledp-graphs/graphs/ego-twitter_adj_f -O ego-twitter_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/gplus_adj_f -O gplus_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/stanford_adj_f -O stanford_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/dblp_adj -O dblp_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/brain_adj -O brain_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/orkut_adj -O orkut_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/livejournal_adj -O livejournal_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/twitter_adj -O twitter_adj
wget -c https://storage.googleapis.com/ledp-graphs/graphs/friendster_adj -O friendster_adj