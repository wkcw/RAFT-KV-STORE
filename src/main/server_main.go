package main

import (
	pb_monkey "chaosmonkey"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "proto"
	"service"
)

func main(){
	service.Port = "9527"
	lis, err := net.Listen("tcp", ":" + service.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	var addrs []string
	addrs = make([]string, 1)
	addrs[0] = "127.0.0.1:9528"
	monkey := service.NewMonkeyService()
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService(addrs, monkey))
	pb_monkey.RegisterChaosMonkeyServer(grpcServer, monkey)
	grpcServer.Serve(lis)
}