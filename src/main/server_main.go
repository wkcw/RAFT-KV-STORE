package main

import (
	pb_monkey "chaosmonkey"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "proto"
	"service"
	"util"
	"os"
)

func main() {
	serverList := util.ServerList{}
	serverList = util.CreateServerList("/Users/wkcw/Desktop/cse223/new/cse223b-RAFT-KV-STORE/src/util/config.xml")
	addr := os.Args[1]
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
	monkey := service.NewMonkeyService(serverList.ServerNum)
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService(serverList, addr, selfServerDescriptor.ServerId, monkey))
	pb_monkey.RegisterChaosMonkeyServer(grpcServer, monkey)
	grpcServer.Serve(lis)
}