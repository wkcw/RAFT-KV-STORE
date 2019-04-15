package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	pb "proto"
	"service"
)

func main(){
	service.Port = "9528"
	lis, err := net.Listen("tcp", ":" + service.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	var addrs []string
	addrs = make([]string, 1)
	addrs[0] = "127.0.0.1:9527"
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService(addrs))
	grpcServer.Serve(lis)
}