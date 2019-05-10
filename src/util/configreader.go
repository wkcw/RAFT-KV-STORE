package util

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	XMLName xml.Name `xml:"config"`
	ServerList ServerList `xml:"servers"`
	RpcTimeout int64 `xml:"rpc_timeout"`
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



func CreateConfig() *Config{
	config := Config{}
	path, _ := filepath.Abs("./src/util/config_sbs.xml")
	//configText, err := ioutil.ReadFile(
	//	"/Users/luxuhui/Desktop/course_work/Distributed_Computing_System/" +
	//		"cse223b-RAFT-KV-STORE/src/util/config_client.xml")

	configText, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("could not parse configure file: %v", err)
	}
	xml.Unmarshal(configText, &config)

	return &config
}


func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}

