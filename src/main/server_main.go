package main

import (
	"flag"
	"net"
	"fmt"
	"log"
	"google.golang.org/grpc"
	pb "proto"
	"service"
)

func main(){
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", service.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService())
	grpcServer.Serve(lis)
}