from collections import defaultdict
from math import ceil
import os
import itertools
from math import comb
import argparse
import random

GRAPH_SIZES = {
    'email-eu-core': 986,
    'wiki': 7115,
    'enron': 36692,
    'brightkite': 58228,
    'ego-twitter': 81306,
    'gplus': 107614,
    'stanford': 281903,
    'dblp': 317080,
    'brain': 784262,
    'orkut': 3072441,
    'livejournal': 4846609,
    'twitter': 41652230,
    'friendster': 65608366
}


def chunk_into_n(lst, n):
  size = ceil(len(lst) / n)
  return [lst[x * size:x * size + size] for x in range(n)]

def calculate_workloads(n, num_process):
    chunk = n // num_process
    extra = n % num_process
    offset = 0
    workloads = []

    for p in range(1, num_process + 1):
        workload = chunk + extra if p == num_process else chunk
        node_ids = list(range(offset, offset + workload))
        workloads.append(node_ids)
        offset += workload

    return workloads

def partition_graph(graph, n, graph_loc):
    processes = n
    graph_directory = f'{graph_loc}{graph}_partitioned_{n}/'
    if not os.path.exists(graph_directory):
        adj_file = f'{graph_loc}{graph}_adj'
        f = open(adj_file, 'r')
        lines = f.readlines()
        lines = [line.strip() for line in lines]
        f.close()
        data = defaultdict(set)
        for l in lines:
            edge = l.split(' ')
            n1 = int(edge[0])
            n2 = int(edge[1])
            data[n1].add(n2)
            if n2 >= 0:
                data[n2].add(n1)
        

        chunked_nodes = calculate_workloads(GRAPH_SIZES[graph.lower()], processes)

        
        os.makedirs(graph_directory, exist_ok=True)
        total_m = 0;
        for i, cn in enumerate(chunked_nodes):
            print("Partition : {0} | Nodes : {1} | ADL : {2}".format(i, len(cn), sum([len(data[n]) for n in cn])))
            total_m += sum([len(data[n]) for n in cn])
            graph_file = graph_directory + '{0}.txt'.format(i)
            with open(graph_file, 'w') as out:
                for node in cn:
                    adjacency_list = data[node]
                    for a in adjacency_list:
                        out.write('{0} {1}\n'.format(node, a))
            out.close()



if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('graph_name', type=str, help="name of graph")
    parser.add_argument('num_workers', type=int, help="number of workers")
    parser.add_argument('graph_loc', type=str, help="location of graph", default="/home/pm886/palmer_scratch/graph-dp-experiments/graphs_new/")
    args = parser.parse_args()
    graph_name = args.graph_name
    partition_graph(graph_name, args.num_workers, args.graph_loc)
    
