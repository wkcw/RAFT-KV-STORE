package service

import (
	"google.golang.org/grpc"
	"log"
	"time"
	"context"
	pb "proto"
)

type connManager struct{
	rpcCaller pb.RaftClient
	conn *grpc.ClientConn
	ctx context.Context
	cancelFunc context.CancelFunc
}

func (cm connManager) gc(){
	cm.conn.Close()
	cm.cancelFunc()
}

func createConnManager(addr string) *connManager {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Connection Failed: %v\n", err)
	}
	// set up a new client
	c := pb.NewRaftClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	retConnManager := &connManager{rpcCaller:c, conn:conn, ctx:ctx, cancelFunc:cancel}
	return retConnManager
}