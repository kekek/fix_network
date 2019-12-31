package hosts

import (
	"wps.ktkt.com/monitor/fix_network/pkg/goodhosts"
)

func NewHosts(hostPath string) (goodhosts.Hosts, error) {
	//osHostsFilePath := os.ExpandEnv(filepath.FromSlash(hostsFilePath))

	hosts := goodhosts.Hosts{Path: hostPath}

	err := hosts.Load()
	if err != nil {
		return hosts, err
	}

	return hosts, nil
}

