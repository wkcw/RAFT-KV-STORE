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

	return config
}


func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
