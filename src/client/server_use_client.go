package client

import (
	pb "proto"
	"fmt"
	"util"
)
type ServerUseClient struct {
	Client
}

func (suc *ServerUseClient) PutTargeted(key string, value string, serverAddr string)(*pb.PutResponse, error){
	cm := createConnManager(serverAddr)
	defer cm.gc()
	fmt.Println("in suc PutTargeted")
	r, err := cm.c.Put(cm.ctx, &pb.PutRequest{Key: key, Value: value})
	fmt.Print("in suc PutTargeted-> r is: ")
	fmt.Println(r)
	return r, err
}

func (suc *ServerUseClient) PutAllOthers(key string, value string)(*pb.PutResponse, error){
	ret := &pb.PutResponse{Ret:pb.ReturnCode_SUCCESS}
	for _, addr := range suc.ServerAddrs{
		fmt.Println("dealing address: "+addr)
		r, err := suc.PutTargeted(key, value, addr)
		fmt.Print(r)
		if r.Ret != pb.ReturnCode_SUCCESS{
			return r, err
		}
	}
	return ret, nil
}


func NewServerUseClient(serverList util.ServerList) *ServerUseClient{
	ret := new(ServerUseClient)
	ret.ServerAddrs = addrs
	return ret
}

