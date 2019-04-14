package service

import (
	"context"
	"sync"
	pb "proto"
)

//var (
//	Port = flag.Int("port", 9527, "The server port")
//)

const (
	Port = 9527
)

type KVService struct{
	lock *sync.RWMutex
	Dict map[string]string
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
	return ret, nil
}

func (kv *KVService) getLocal(key string) (string, error){
	kv.lock.RLock()
	defer kv.lock.RUnlock()
	if value, ok := kv.Dict[key]; ok {
		return value, nil
	}else{
		return value, &KeyError{key}
	}
}

func (kv *KVService) putLocal(key string, data string){
	kv.lock.Lock()
	defer kv.lock.Unlock()
	kv.Dict[key] = data
}

