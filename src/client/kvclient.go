package client

import (
	"flag"
	"math/rand"
	"google.golang.org/grpc"
	"log"
	"time"
	"context"
	pb "proto"
)

var (
	ServerAddr = flag.String("server_addr", "127.0.0.1:9527", "The server address in the format of host:port")
)


type Client struct{
	serverAddrs []string
}

type connManager struct{
	c pb.KeyValueStoreClient
	conn *grpc.ClientConn
	ctx context.Context
	cancelFunc context.CancelFunc
}

func (cm connManager) gc(){
	cm.conn.Close()
	cm.cancelFunc()
}

func NewClient(addrs []string) *Client{
	ret := new(Client)
	ret.serverAddrs = addrs;
	return ret
}

func createConnManager(addr string) *connManager {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Connection Failed: %v", err)
	}
	// set up a new client
	c := pb.NewKeyValueStoreClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	retConnManager := &connManager{c:c, conn:conn, ctx:ctx, cancelFunc:cancel}
	return retConnManager
}

func (client *Client) pickRandomServer() string{
	return client.serverAddrs[rand.Intn(len(client.serverAddrs))]
}

func (client *Client) Put(key string, value string)(*pb.PutResponse, error){
	serverAddr := client.pickRandomServer()
	cm := createConnManager(serverAddr)
	defer cm.gc()
	r, err := cm.c.Put(cm.ctx, &pb.PutRequest{Key: key, Value: value})
	return r, err
}

func (client *Client) Get(key string)(*pb.GetResponse, error){
	serverAddr := client.pickRandomServer()
	cm := createConnManager(serverAddr)
	defer cm.gc()
	r, err := cm.c.Get(cm.ctx, &pb.GetRequest{Key: key})
	return r, err
}