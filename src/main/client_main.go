package main

import (
	"client"
	"log"
	"util"
)

var (
	ServerAddrs = []string{"127.0.0.1:9527"}
	ServerAddr = "127.0.0.1:9527"
	operation, key, value string
)

func main() {
	serverList := util.ServerList{}
	serverList = util.CreateServerList("/Users/wkcw/Desktop/cse223/new/cse223b-RAFT-KV-STORE/src/util/config.xml")
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	//for {
		//fmt.Scanln(&operation, &key, &value)
		operation = "put"
		if (operation == "put") {
			r, err := client.PutAndBroadcast(key, value)
			if err != nil {
				log.Fatalf("could not put: %v", err)
			}
			log.Printf("Return code: %s", r.Ret)
		}

		if (operation == "get") {
			r1, err1 := client.Get(key)
			if err1 != nil {
				log.Fatalf("could not get: %v", err1)
			}
			log.Printf("Value: %s", r1.Value)
		}
	//}
}