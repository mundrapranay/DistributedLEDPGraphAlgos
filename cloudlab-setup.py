from pssh.clients import ParallelSSHClient


def get_hosts(hostFileName):
    hostFile = open(hostFileName, 'r')
    lines = hostFile.readlines()
    hosts = [line.strip() for line in lines]
    hostFile.close()
    return hosts


def test_command(hosts):
    client = ParallelSSHClient(hosts)
    output = client.run_command('date')
    for host_output in output:
        for line in host_output.stdout:
            print(line)
        exit_code = host_output.exit_code


if __name__ == '__main__':
    hosts = get_hosts('cloudlab')
    test_command(hosts)
