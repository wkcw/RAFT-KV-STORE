package main

import (
	"log"
	"fmt"
	"client"
)

var (
	ServerAddrs = []string{"127.0.0.1:9527"}
	ServerAddr = "127.0.0.1:9527"
	operation, key, value string
)

func main() {
	// Set up a client to a set of servers
	client := client.NewClient(ServerAddrs)

	for {
		fmt.Scanln(&operation, &key, &value)
		if (operation == "put") {
			r, err := client.Put(key, value)
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
	}
}