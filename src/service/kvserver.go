package service

import (
	pb_monkey "chaosmonkey"
	"client"
	"context"
	"fmt"
	pb "proto"
	"sync"
	"util"
	"math/rand"
	"google.golang.org/grpc/peer"
	"net"
	"log"
	"strconv"
)



type KVService struct{
	lock *sync.RWMutex
	dict map[string]string
	clientToOthers *client.ServerUseClient
	monkey *MonkeyService
	selfAddr string
	selfID int
	addrToID map[string]int


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
	//TODO
	if kv.monkey != nil{
		pr, ok := peer.FromContext(ctx)
		if !ok {
			log.Fatalf("[getClinetIP] invoke FromContext() failed")
		}
		if pr.Addr == net.Addr(nil) {
			log.Fatalf("[getClientIP] peer.Addr is nil")
		}
		senderAddr := pr.Addr.String()
		fmt.Printf(senderAddr)
		if !kv.whetherToDrop(req.SelfID){
			e := new(PacketLossError)
			e.Msg = "you didnt pass ChaosMonkey"
			return nil, e
		}
	}
	//TODO
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
	for _, v := range kv.monkey.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
}

func (kv *KVService) PutAndBroadcast(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	fmt.Println("in PutAndBroadcast")
	key, val := req.Key, req.Value
	kv.putLocal(key, val)
	kv.clientToOthers.PutAllOthers(key, val);
	ret := &pb.PutResponse{Ret: pb.ReturnCode_SUCCESS}
	return ret, nil
}

func (kv *KVService)putOtherServers(key string, data string){
	kv.clientToOthers.PutAllOthers(key, data)
}


func NewKVService(serverList util.ServerList, selfAddr string, selfID int, monkey *MonkeyService) *KVService{
	ret := &KVService{lock:new(sync.RWMutex), dict:make(map[string]string), clientToOthers:client.NewServerUseClient(serverList, selfAddr, selfID), monkey: monkey}
	ret.selfAddr = selfAddr
	ret.selfID = selfID
	ret.addrToID = make(map[string]int)
	for _, sd := range serverList.Servers{
		ret.addrToID[sd.Host+":"+sd.Port] = sd.ServerId
	}
	return ret
}

func NewMonkeyService(n int) *MonkeyService {
	return &MonkeyService{matrix: make([][]float32, n, n)}
}

func (kv *KVService) whetherToDrop(senderID string) bool{
	fmt.Printf("senderID is %s, self ID is %d\n", senderID, kv.selfID)
	randNum := rand.Float32()
	intSenderID, _ := strconv.Atoi(senderID)
	probInMat := kv.monkey.matrix[intSenderID][kv.selfID]
	fmt.Println("Num in Mat is %f", probInMat)
	fmt.Println("Generated RandNum is %f", randNum)

	return probInMat<randNum //if true message received
}