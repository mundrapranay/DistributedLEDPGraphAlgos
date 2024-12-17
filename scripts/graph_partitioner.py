from collections import defaultdict
from math import ceil
import os
import itertools
from math import comb
import argparse
import random

GRAPH_SIZES = {
    'hua_ctr' : 14081816,
    'livejournal' : 4846609,
    'hua_stackoverflow' : 2584164,
    'hua_usa' : 23947347,
    'hua_youtube' : 1138499,
    'orkut' : 3072441,
    'gplus' : 107614,
    'imdb' : 896308,
    'complete_small' : 1000,
    'random_gen_2' : 2500,
    'big_random' : 100000,
    'dblp' : 317080,
    'small-graph' : 17,
    'random-graph' : 1000,
    'twitter' : 41652230,
    'brain' : 784262,
    'friendster' : 65608366,
    'wiki' : 7115,
    'enron' : 36692,
    'brightkite' : 58228,
    'imdb3' : 892457
}



def read_graph_from_file(filename):
    graph = {}
    with open(filename, 'r') as file:
        for line in file:
            u, v = line.strip().split()
            if u not in graph:
                graph[u] = []
            if v not in graph:
                graph[v] = []
            graph[u].append(v)
            graph[v].append(u)  # Assuming an undirected graph
    return graph


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

def partition_graph(graph, n):
    processes = n
    graph_directory = '/home/pranaymundra/graph-dp-experiments/graphs_new/{0}_partitioned_{1}/'.format(graph.lower(), n)
    if not os.path.exists(graph_directory):
        f = open('/home/pranaymundra/graph-dp-experiments/graphs_new/{0}_adj'.format(graph), 'r')
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



def load_graph(graph):
    f = open('/home/ubuntu/graph-dp-experiments/graphs/{0}_adj'.format(graph), 'r')
    lines = f.readlines()
    del lines[0]
    lines = [line.strip() for line in lines]
    f.close()
    data = defaultdict(list)
    for l in lines:
        edge = l.split(' ')
        n1 = int(edge[0])
        n2 = int(edge[1])
        data[n1].append(n2)
        data[n2].append(n1)
    return data



def random_sample_graph(graph, sample_size):
    sampled_nodes = random.sample([i for i in range(GRAPH_SIZES[graph])], sample_size)
    # print(sampled_nodes[:30])
    data = load_graph(graph)
    edges = 0
    nodes = 0
    graph_file = '/home/ubuntu/graph-dp-experiments/graphs/{0}_sampled_{1}'.format(graph, sample_size)
    with open(graph_file, 'w') as out:
        for node in sampled_nodes:
            nodes += 1
            adjacency_list = data[node]
            adjacency_list_sampled = set(sampled_nodes).intersection(set(adjacency_list))
            for a in adjacency_list_sampled:
                edges += 1
                out.write('{0} {1}\n'.format(node, a))
    out.close()
    print('Nodes: {0}\t Edges: {1}\n'.format(nodes, edges))


def reindex_and_generate_subgraph(graphname, k):

    graph = load_graph(graphname)
    # Step 1: Randomly sample k nodes from the graph
    sampled_nodes = random.sample(list(graph.keys()), k)
    
    # Step 2: Create a mapping for the nodes from original to [0, k)
    node_mapping = {node: idx for idx, node in enumerate(sampled_nodes)}
    
    # Step 3: Generate re-indexed subgraph
    subgraph = {}
    for node in sampled_nodes:
        # Get the new index for the node
        new_node_idx = node_mapping[node]
        # Include edges that connect to other nodes within the sampled set, using the new indexing
        subgraph[new_node_idx] = [node_mapping[neighbor] for neighbor in graph[node] if neighbor in sampled_nodes]


    edges = 0
    graph_file = '/home/ubuntu/graph-dp-experiments/graphs/{0}_sampled_{1}'.format(graphname, k)
    with open(graph_file, 'w') as out:
        for node in subgraph:
            adjacency_list = subgraph[node]
            edges += len(adjacency_list)
            for a in adjacency_list:
                out.write('{0} {1}\n'.format(node, a))
    out.close()
    print('Nodes: {0}\t Edges: {1}\n'.format(len(subgraph), edges))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('graph_name', type=str, help="name of graph")
    parser.add_argument('num_workers', type=int, help="number of workers")
    args = parser.parse_args()
    graph_name = args.graph_name
    partition_graph(graph_name, args.num_workers)
    
