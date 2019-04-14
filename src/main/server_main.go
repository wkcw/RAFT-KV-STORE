package main

import (
	"flag"
	"net"
	"fmt"
	"log"
	"os"
	"google.golang.org/grpc"
	pb "proto"
	"service"
)

func main(){
	service.Port = os.Args[1]
	lis, err := net.Listen("tcp", ":" + service.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(grpcServer, service.NewKVService())
	grpcServer.Serve(lis)
}