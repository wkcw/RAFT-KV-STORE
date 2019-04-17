package client

import (
	pb "proto"
	"util"
	"fmt"
	"strconv"
)
type ServerUseClient struct {
	Client
	selfAddr string
	selfID int
}

func (suc *ServerUseClient) PutTargeted(key string, value string, serverAddr string)(*pb.PutResponse, error){
	cm := createConnManager(serverAddr)
	defer cm.gc()
	r, err := cm.c.Put(cm.ctx, &pb.PutRequest{Key: key, Value: value, SelfID: strconv.Itoa(suc.selfID)})
	return r, err
}

func (suc *ServerUseClient) PutAllOthers(key string, value string)(*pb.PutResponse, error){
	ret := &pb.PutResponse{Ret:pb.ReturnCode_SUCCESS}
	for _, sd := range suc.ServerList.Servers{
		if sd.Host+":"+sd.Port != suc.selfAddr{
			r, err := suc.PutTargeted(key, value, sd.Host+":"+sd.Port)
			if err!=nil{
				fmt.Println("Msg sent to "+sd.Host+":"+sd.Port+" was dropped")
				return nil, err
			} else if r.Ret != pb.ReturnCode_SUCCESS{
				return r, err
			}
		}
	}
	return ret, nil
}


func NewServerUseClient(serverList util.ServerList, selfAddr string, selfID int) *ServerUseClient{
	ret := new(ServerUseClient)
	ret.ServerList = serverList
	ret.selfAddr = selfAddr
	ret.selfID = selfID
	return ret
}

