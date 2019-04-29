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
	HeartbeatInterval int64 `xml:"heartbeat_interval"`
	ServerList ServerList `xml:"servers"`
	ID string `xml:"ID"`
	TimeoutUpperBound int64 `xml:"timeout_upper_bound"`
	TimeoutLowerBound int64 `xml:"timeout_lower_bound"`
}

type Matrix struct {
	XMLName xml.Name `xml:"matrix"`
	Filepath string `xml:"filepath,attr"`
}

type ServerList struct {
	XMLName xml.Name `xml:"servers"`
	ServerNum int `xml:"nums,attr"`
	Servers []Server `xml:"server"`
}

type Server struct {
	XMLName xml.Name `xml:"server"`
	ServerId int `xml:"serverId"`
	Addr string `xml:"addr"`
}


func CreateConfig()  raftConfig{
	config := raftConfig{}

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
