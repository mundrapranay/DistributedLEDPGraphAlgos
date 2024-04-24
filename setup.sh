#!/bin/sh
cd
mkdir results
mkdir graph-dp-experiments
cd graph-dp-experiments
mkdir graphs
cd
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
cd localgraph-dp/
go mod tidy
cd