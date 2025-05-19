import numpy as np 
import statistics
import math 


TCOUNT = {
    'wiki' : 608387,
    'enron' : 727044,
    'brightkite' : 494728,
    'gplus' : 123652925160,
    'gplus' : 1073677742,
    'dblp' : 2224385.00000000,
    'stanford' : 11329473,
    'email-eu-core' : 105461,
    'ego-twitter' : 13082506
}

KCORE_FILE = "{0}_rr-kcore_0.80_1_1_8_{1}_80_1.00_rr_baseline.txt"
TCOUNT_FILE = "{0}_rr-tcount_0.80_1_1_8_{1}_80_1.00_rr_baseline.txt"

def get_max_approx_index(pairs):
    # Define a custom key function to extract the 'approx' value from each pair
    def get_approx(pair):
        return pair[1]

    # Find the pair with the maximum 'approx' value using the custom key function
    max_pair = max(pairs, key=get_approx)

    # Find the index of the max_pair in the original list
    max_index = pairs.index(max_pair)

    return max_index


# def get_core_numbers(file):
#     f = open('/home/pm886/palmer_scratch/results/{0}'.format(file), 'r')
#     lines = f.readlines()
#     f.close()
#     lines = [line.strip().split(':') for line in lines]
#     estimated_core_numbers = []
#     for line in lines:
#         cn = float(line[1].strip())
#         estimated_core_numbers.append(cn)
#     return estimated_core_numbers

def get_core_numbers(filename):
    path = f'/home/pm886/palmer_scratch/results/{filename}'
    with open(path, 'r') as f:
        pairs = []
        for line in f:
            line = line.strip()
            if not line:
                continue
            node_str, core_str = line.split(':', 1)
            try:
                node = int(node_str)
            except ValueError:
                # if node IDs aren’t integers, leave as string
                node = node_str
            core = float(core_str)
            pairs.append((node, core))

    # sort by node (int or string)
    pairs.sort(key=lambda x: x[0])

    # return just the core numbers, in node-ID order
    return [core for _, core in pairs]


def get_ground_truth(graph):
    f = open('/home/pm886/palmer_scratch/graph-dp-experiments/ground_truth/{0}'.format(graph), 'r')
    lines = f.readlines()
    f.close()
    lines = [line.strip().split(' ') for line in lines]
    core_numbers = []
    for line in lines:
        try:
            cn = float(line[1].strip())
            core_numbers.append(cn)
        except ValueError:
            continue
    return core_numbers

def get_triangles(file):
    f = open('/home/pm886/palmer_scratch/results/{0}'.format(file), 'r')
    lines = f.readlines()
    f.close()
    return float(lines[0].strip().split(':')[1])


def get_kcore_data():
    avg_approx = []
    eighty_approx = []
    ninefive_approx = []
    graphs = ['email-eu-core', 'wiki', 'enron', 'brightkite', 'ego-twitter', 'gplus', 'stanford', 'dblp' ]
    for graph in graphs:
        core_numbers = get_ground_truth(graph)
        for run_id in range(5):
            avg_approx_l = []
            eighty_approx_l = []
            ninefive_approx_l = []
            file = KCORE_FILE.format(graph, run_id)
            approx_core_numbers = get_core_numbers(file)
            approximation_factor = np.array([float(max(s,t)) / max(1, min(s, t)) for s,t in zip(core_numbers, approx_core_numbers)])
            avg_approx_l.append(statistics.mean(approximation_factor))
            eighty_approx_l.append(np.percentile(approximation_factor, 80))
            ninefive_approx_l.append(np.percentile(approximation_factor, 95))
            
        avg_approx.append(statistics.mean(avg_approx_l))
        eighty_approx.append(statistics.mean(eighty_approx_l))
        ninefive_approx.append(statistics.mean(ninefive_approx_l))

    print('\t'.join(graphs))
    print('\t'.join(f"{x:.3f}" for x in avg_approx))
    print('\t'.join(f"{x:.3f}" for x in eighty_approx))
    print('\t'.join(f"{x:.3f}" for x in ninefive_approx))

def calculate_confidence_interval(data, cf_level):
    mean = np.mean(data)
    stdev = np.std(data)
    margin_of_error = cf_level * stdev / math.sqrt(len(data))
    return mean, mean - margin_of_error, mean + margin_of_error

def get_tcount_data():
    rel_error = []
    avg_approx = []
    graphs = ['email-eu-core', 'wiki', 'enron', 'brightkite']
    rel_error_bounds = []
    for graph in graphs:
        tcount = TCOUNT[graph]
        erro_l_count = []
        for run_id in range(5):
            rel_error_l = []
            avg_approx_l = []
            file = TCOUNT_FILE.format(graph, run_id)
            try:
                approx_tcount = get_triangles(file)
                avg_approx_l.append(float(max(tcount,approx_tcount)) / max(1, min(tcount, approx_tcount)))
                rel_error_l.append(float(abs(approx_tcount - tcount) / tcount))
                erro_l_count.append(float(abs(approx_tcount - tcount) / tcount))
            except FileNotFoundError:
                continue
        mean, lower, upper = calculate_confidence_interval(erro_l_count, 1.96)
        rel_error.append(mean)
        rel_error_bounds.append([lower, upper])
        avg_approx.append(statistics.mean(avg_approx_l))

    print('\t'.join(graphs))
    print('\t'.join(f"{x:.3f}" for x in avg_approx))
    print('\t'.join(f"{x:.3f}" for x in rel_error))
    print(rel_error_bounds)


def kcore_results_disc():
    approx_core_numbers = get_core_numbers('wiki_distributedKV.txt')
    core_numbers = get_ground_truth('wiki')
    approximation_factor = np.array([float(max(s,t)) / max(1, min(s, t)) for s,t in zip(core_numbers, approx_core_numbers)])
    print(statistics.mean(approximation_factor))
    print(np.percentile(approximation_factor, 80))
    print(np.percentile(approximation_factor, 95))
    # print(len(approx_core_numbers))

if __name__ == '__main__':
    # get_tcount_data()
    # get_kcore_data()
    kcore_results_disc()