//package util

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"encoding/xml"
)

type ServerList struct {
	XMLName xml.Name `xml:"servers"`
	ServerNum string `xml:"nums,attr"`
	Servers []Server `xml:"server"`
}

type Server struct {
	XMLName xml.Name `xml:"server"`
	ServerId string `xml:"serverId"`
	Host string `xml:"host"`
	Port string `xml:"port"`
}


func createServerList(filename string)  *ServerList{
	sList := ServerList{}

	config, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("could not parse configure file: %v", err)
		os.Exit(9)
	}

	fmt.Println(string(config))

	xml.Unmarshal(config, &sList)

	return &sList
}

func main()  {
	test := createServerList("config.xml")

	fmt.Println(test)
}