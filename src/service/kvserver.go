package service

import (
	pb_monkey "chaosmonkey"
	"client"
	"context"
	"fmt"
	pb "proto"
	"sync"
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

type MonkeyService struct {
	matrix [][]float32
}

func (s *MonkeyService) UploadMatrix(ctx context.Context, req *pb_monkey.ConnMatrix) (*pb_monkey.Status, error) {
	rows := req.GetRows()
	for i, v := range rows {
		s.matrix[i] = v.GetVals()
	}
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
}

func (s *MonkeyService) UpdateValue(ctx context.Context, req *pb_monkey.MatValue) (*pb_monkey.Status, error) {
	row := req.Row
	col := req.Col
	value := req.Val
	s.matrix[row][col] = value
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
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

func NewMonkeyService() *MonkeyService {
	return &MonkeyService{matrix: make([][]float32, 5, 5)}
}
func (kv *KVService) SetOtherAddrs(addrs []string){
	kv.clientToOthers = client.NewServerUseClient(addrs)
}