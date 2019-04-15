package service

import (
	"context"
	"sync"
	pb "proto"
	"client"
	"fmt"
)

//var (
//	Port = flag.Int("port", 9527, "The server port")
//)

var (
	Port string
)

type KVService struct{
	lock *sync.RWMutex
	dict map[string]string
	clientToOthers *client.ServerUseClient
}


//Get(context.Context, *GetRequest) (*GetResponse, error)
func (kv *KVService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error){
	key := req.Key
	data, e := kv.getLocal(key)
	if e == nil {
		ret := &pb.GetResponse{Value: data, Ret: pb.ReturnCode_SUCCESS}
		return ret, nil
	}else{
		return nil, e
	}
}

func (kv *KVService) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	key, val := req.Key, req.Value
	kv.putLocal(key, val)
	ret := &pb.PutResponse{Ret: pb.ReturnCode_SUCCESS}
	fmt.Print("in Put impl")
	fmt.Println(ret)
	return ret, nil
}

func (kv *KVService) getLocal(key string) (string, error){
	kv.lock.RLock()
	defer kv.lock.RUnlock()
	if value, ok := kv.dict[key]; ok {
		return value, nil
	}else{
		return value, &KeyError{key}
	}
}

func (kv *KVService) putLocal(key string, data string){
	kv.lock.Lock()
	defer kv.lock.Unlock()
	kv.dict[key] = data
}

func (kv *KVService) PutAndBroadcast(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	key, val := req.Key, req.Value
	kv.putLocal(key, val)
	kv.clientToOthers.PutAllOthers(key, val);
	ret := &pb.PutResponse{Ret: pb.ReturnCode_SUCCESS}
	return ret, nil
}

func (kv *KVService)putOtherServers(key string, data string){
	kv.clientToOthers.PutAllOthers(key, data)
}


func NewKVService(addrs []string) *KVService{
	ret := &KVService{lock:new(sync.RWMutex), dict:make(map[string]string), clientToOthers:client.NewServerUseClient(addrs)}
	return ret
}

func (kv *KVService) SetOtherAddrs(addrs []string){
	kv.clientToOthers = client.NewServerUseClient(addrs)
}