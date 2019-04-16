package client

import (
	"math/rand"
	"google.golang.org/grpc"
	"log"
	"time"
	"context"
	pb "proto"
	"fmt"
	"util"
)


type Client struct{
	ServerList util.ServerList
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

func NewClient(serverList util.ServerList) *Client{
	ret := new(Client)
	ret.ServerList = serverList;
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
	randNum := rand.Intn(client.ServerList.ServerNum)
	sd := client.ServerList.Servers[randNum]
	port := sd.Port
	ip := sd.Host
	fmt.Println(ip+":"+port)
	return ip+":"+port
}

func (client *Client) PutAndBroadcast(key string, value string)(*pb.PutResponse, error){
	serverAddr := client.pickRandomServer()
	r, err := client.PutTargetedAndBroadcast(key, value, serverAddr)
	return r, err
}

func (client *Client) PutTargetedAndBroadcast(key string, value string, serverAddr string)(*pb.PutResponse, error){
	fmt.Println("in PutTargetedAndBroadcast: "+serverAddr)
	cm := createConnManager(serverAddr)
	fmt.Println(cm)
	defer cm.gc()
	r, err := cm.c.PutAndBroadcast(cm.ctx, &pb.PutRequest{Key: key, Value: value})
	return r, err
}

func (client *Client) Get(key string)(*pb.GetResponse, error){
	serverAddr := client.pickRandomServer()
	fmt.Println("From server "+serverAddr+" got:")
	r, err := client.GetTargeted(key, serverAddr)
	return r, err
}

func (client *Client) GetTargeted(key string, serverAddr string)(*pb.GetResponse, error){
	cm := createConnManager(serverAddr)
	defer cm.gc()
	r, err := cm.c.Get(cm.ctx, &pb.GetRequest{Key: key})
	return r, err
}

