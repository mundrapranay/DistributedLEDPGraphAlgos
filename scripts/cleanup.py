import shutil
import argparse

def cleanup(graph, n, graph_loc):
    graph_directory = f'{graph_loc}{graph}_partitioned_{n}/'
    shutil.rmtree(graph_directory, ignore_errors=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('graph_name', type=str, help="name of graph")
    parser.add_argument('num_workers', type=int, help="number of workers")
    parser.add_argument('graph_loc', type=str, help="location of graph", default="../graph-dp-experiments/graphs/")
    args = parser.parse_args()
    graph_name = args.graph_name
    cleanup(graph_name, args.num_workers, args.graph_loc)