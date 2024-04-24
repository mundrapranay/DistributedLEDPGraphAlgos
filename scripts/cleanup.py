import shutil
import argparse

def cleanup(graph, n):
    graph_directory = '/home/ubuntu/graph-dp-experiments/graphs/{0}_partitioned_{1}/'.format(graph.lower(), n)
    shutil.rmtree(graph_directory, ignore_errors=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('graph_name', type=str, help="name of graph")
    parser.add_argument('num_workers', type=int, help="number of workers")
    args = parser.parse_args()
    graph_name = args.graph_name
    cleanup(graph_name, args.num_workers)