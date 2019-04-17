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
	config := util.CreateConfig()
	serverList := config.ServerList
	addr := os.Args[1]
	log.Println("I am server: "+addr)
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