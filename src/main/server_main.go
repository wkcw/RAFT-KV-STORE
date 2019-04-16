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

func main(){
	serverList := &util.ServerList{}
	serverList = util.CreateServerList("/Users/wkcw/Desktop/cse223/new/cse223b-RAFT-KV-STORE")
	addrPort := os.Args[1]
	var selfServerDescriptor util.Server
	for _, server := range serverList.Servers{
		if server.Host+":"+server.Port == addrPort{
			selfServerDescriptor = server
	}

	lis, err := net.Listen("tcp", ":" + selfServerDescriptor.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	//addrs[0] = "127.0.0.1:9527"
	monkey := service.NewMonkeyService()
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService(serverList, monkey))
	pb_monkey.RegisterChaosMonkeyServer(grpcServer, monkey)
	grpcServer.Serve(lis)
}