//package util

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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


func CreateServerList(filename string)  ServerList{
	sList := ServerList{}

	config, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("could not parse configure file: %v", err)
	}
	xml.Unmarshal(config, &sList)

	return sList
}


func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}

