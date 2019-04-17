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
	Matrix Matrix `xml:"matrix"`
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
	Host string `xml:"host"`
	Port string `xml:"port"`
}


func CreateConfig()  Config{
	config := Config{}

	configText, err := ioutil.ReadFile("/Users/wkcw/Desktop/cse223/garbage/cse223b-RAFT-KV-STORE/src/util/config.xml")
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
