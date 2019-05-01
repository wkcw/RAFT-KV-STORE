package service

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type raftConfig struct {
	XMLName xml.Name `xml:"config"`
	selfAddr string `xml:"self_addr"`
	heartbeatInterval int64 `xml:"heartbeat_interval"`
	serverList ServerList `xml:"servers"`
	ID string `xml:"ID"`
	electionTimeoutUpperBound int64 `xml:"election_timeout_upper_bound"`
	electionTimeoutLowerBound int64 `xml:"election_timeout_lower_bound"`
	rpcTimeout int64 `xml:"rpc_timeout"`
}

type Matrix struct {
	XMLName xml.Name `xml:"matrix"`
	Filepath string `xml:"filepath,attr"`
}

type ServerList struct {
	XMLName xml.Name `xml:"servers"`
	serverNum int `xml:"nums,attr"`
	servers []Server `xml:"server"`
}

type Server struct {
	XMLName xml.Name `xml:"server"`
	ServerId int `xml:"serverId"`
	addr string `xml:"addr"`
}


func createConfig() *raftConfig{
	config := &raftConfig{}

	configText, err := ioutil.ReadFile("../util/config.xml")
	if err != nil {
		log.Fatalf("could not parse configure file: %v", err)
	}
	xml.Unmarshal(configText, &config)

	selfIndex := -1

	for i, server := range config.serverList.servers {
		if server.addr == config.selfAddr {
			selfIndex = i
		}
	}

	if selfIndex == -1 {
		log.Fatalf("Address does not match with Configuration.")
	}

	config.serverList.servers = append(config.serverList.servers[0:selfIndex],
		config.serverList.servers[selfIndex + 1:]...)

	return config
}


func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
