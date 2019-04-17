package main

import (
	pb_monkey "chaosmonkey"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "proto"
	"service"
	"util"
)

func main() {
	serverList := util.ServerList{}
	serverList = util.CreateServerList("/Users/luxuhui/Desktop/课程/Distributed Computing&System/cse223b-RAFT-KV-STORE/src/util/config.xml")
	addr := os.Args[1]
	fmt.Println(addr)

	var selfServerDescriptor util.Server
	for _, server := range serverList.Servers {
		if server.Host+":"+server.Port == addr {
			selfServerDescriptor = server
		}
	}

	lis, err := net.Listen("tcp", ":"+selfServerDescriptor.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	monkey := service.NewMonkeyService()
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService(serverList, addr, monkey))
	pb_monkey.RegisterChaosMonkeyServer(grpcServer, monkey)
	grpcServer.Serve(lis)
}
