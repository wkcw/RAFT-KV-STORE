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
	XMLName                   xml.Name   `xml:"config"`
	SelfAddr                  string     `xml:"self_addr"`
	HeartbeatInterval         int64      `xml:"heartbeat_interval"`
	ServerList                ServerList `xml:"servers"`
	ID                        string     `xml:"ID"`
	ElectionTimeoutUpperBound int64      `xml:"election_timeout_upper_bound"`
	ElectionTimeoutLowerBound int64      `xml:"election_timeout_lower_bound"`
	RpcTimeout                int64      `xml:"rpc_timeout"`
}


type ServerList struct {
	XMLName   xml.Name `xml:"servers"`
	ServerNum int      `xml:"nums,attr"`
	Servers   []Server `xml:"server"`
}

type Server struct {
	XMLName xml.Name `xml:"server"`
	ServerId int `xml:"serverId"`
	Addr string `xml:"addr"`
}



func createConfig() *raftConfig{
	config := raftConfig{}

	configText, err := ioutil.ReadFile(
		"/Users/wkcw/Desktop/cse223/garbage/node2/src/util/config_local.xml")
	if err != nil {
		log.Fatalf("could not parse configure file: %v", err)
	}
	xml.Unmarshal(configText, &config)

	selfIndex := -1

	for i, server := range config.ServerList.Servers {
		if server.Addr == config.SelfAddr {
			selfIndex = i
		}
	}

	if selfIndex == -1 {
		log.Fatalf("Address does not match with Configuration.")
	}

	config.ServerList.Servers = append(config.ServerList.Servers[0:selfIndex],
		config.ServerList.Servers[selfIndex + 1:]...)

	return &config
}


func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
