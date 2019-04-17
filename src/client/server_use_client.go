package client

import (
	pb "proto"
	"util"
	"fmt"
	"strconv"
	"io"
	"log"
	"time"
)
type ServerUseClient struct {
	Client
	selfAddr string
	selfID int
}

type PacketLossError struct{
	Msg string
}

func (e *PacketLossError) Error() string{
	return fmt.Sprintf("Packet Loss Error -> msg: %s", e.Msg)
}


func (suc *ServerUseClient) PutTargeted(key string, value string, serverAddr string)(*pb.PutResponse, error){
	cm := createConnManager(serverAddr)
	defer cm.gc()
	r, err := cm.c.Put(cm.ctx, &pb.PutRequest{Key: key, Value: value, SelfID: strconv.Itoa(suc.selfID)})
	if r==nil{log.Printf("Timeout got nil ret")}
	return r, err
}

func (suc *ServerUseClient) PutTargetedToGetStreamResponse(key string, value string, serverAddr string)(*pb.PutResponse, error){
	t1 := time.Now()
	cm := createConnManager(serverAddr)
	defer cm.gc()
	stream, _ := cm.c.PutToGetStreamResponse(cm.ctx, &pb.PutRequest{Key: key, Value: value, SelfID: strconv.Itoa(suc.selfID)})

		ret, err := stream.Recv()
		if err == io.EOF {
			log.Println("Packet Lost")
			t2 := time.Now()
			fmt.Println("Time Elapsed When packet loss" + t2.Sub(t1).String())
			packetLossError := new(PacketLossError)
			packetLossError.Msg = "Packet Lost"
			return nil, packetLossError
		}
		if err != nil {
			log.Fatalf("stream.Recv() error is not nil")
		}
		log.Println(ret.Ret)
		return ret, nil
}

func (suc *ServerUseClient) PutAllOthers(key string, value string)(*pb.PutResponse, error){
	ret := &pb.PutResponse{Ret:pb.ReturnCode_SUCCESS}
	for _, sd := range suc.ServerList.Servers{
		if sd.Host+":"+sd.Port != suc.selfAddr{
			log.Printf("Broadcasting to server %s\n", sd.Host+":"+sd.Port)
			r, err := suc.PutTargeted(key, value, sd.Host+":"+sd.Port)
			if err!=nil{
				log.Println("Msg sent to "+sd.Host+":"+sd.Port+" was dropped")
				//return nil, err
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

