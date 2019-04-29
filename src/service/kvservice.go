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
	"google.golang.org/grpc"
)



type KVService struct{
	dictLock       *sync.RWMutex
	dict           map[string]string
	clientToOthers *client.ServerUseClient
	monkey         *MonkeyService
	selfAddr       string
	selfID         int
	addrToID       map[string]int
	idToAddr       map[int]string
	raft           *RaftService
	appendChan     chan entry
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
		ret := &pb.GetResponse{Value: data, Ret: pb.ReturnCode_FAILURE}
		return ret, e
	}
}

//func (kv *KVService) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
//	if kv.monkey != nil{
//		pr, ok := peer.FromContext(ctx)
//		if !ok {
//			log.Fatalf("[getClinetIP] invoke FromContext() failed")
//		}
//		if pr.Addr == net.Addr(nil) {
//			log.Fatalf("[getClientIP] peer.Addr is nil")
//		}
//		senderAddr := pr.Addr.String()
//		fmt.Printf(senderAddr)
//		if !kv.notToDrop(req.SelfID){
//			e := new(PacketLossError)
//			e.Msg = "you didnt pass ChaosMonkey"
//			time.Sleep(2000 * time.Millisecond)
//			timeoutRet := &pb.PutResponse{Ret:pb.ReturnCode_SUCCESS}
//			return timeoutRet, e
//		}
//	}
//	key, val := req.Key, req.Value
//	kv.putLocal(key, val)
//	log.Printf("I received a Broadcast request with Key: %s, Value: %s", key, val)
//	ret := &pb.PutResponse{Ret: pb.ReturnCode_SUCCESS}
//	return ret, nil
//
//}

func (kv *KVService) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	//todo
	//if I am not leader, tell client leader ID and Address

	key, val := req.Key, req.Value
	applyChan := make(chan bool)
	logEntry := entry{op:"put", key:key, val:val, term:-1, applyChan:applyChan}
	kv.appendChan <- logEntry
	applyStatus := <- applyChan
	ret := &pb.PutResponse{}
	if applyStatus{
		kv.putLocal(key, val)
		log.Printf("Successfully applied a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_SUCCESS
	}else{
		log.Printf("Failed to apply a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_FAILURE
	}
	return ret, nil

}

func (kv *KVService) getLocal(key string) (string, error){
	kv.dictLock.RLock()
	defer kv.dictLock.RUnlock()
	if value, ok := kv.dict[key]; ok {
		return value, nil
	}else{
		return value, &KeyError{key}
	}
}

func (kv *KVService) putLocal(key string, data string){
	kv.dictLock.Lock()
	defer kv.dictLock.Unlock()
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


func NewKVService(serverList util.ServerList, selfAddr string, selfID int, monkey *MonkeyService) *KVService{
	kv := &KVService{dictLock: new(sync.RWMutex), dict:make(map[string]string), clientToOthers:client.NewServerUseClient(serverList, selfAddr, selfID), monkey: monkey}
	kv.selfAddr = selfAddr
	kv.selfID = selfID
	kv.addrToID = make(map[string]int)
	for _, sd := range serverList.Servers{
		kv.addrToID[sd.Host+":"+sd.Port] = sd.ServerId
	}
	appendChan := make(chan entry)
	kv.appendChan = appendChan
	raft := NewRaftService(kv.appendChan)
	kv.raft = raft
	return kv
}

func NewMonkeyService(n int) *MonkeyService {
	return &MonkeyService{matrix: make([][]float32, n, n)}
}

func (kv *KVService) notToDrop(senderID string) bool{
	log.Printf("senderID is %s, self ID is %d\n", senderID, kv.selfID)
	randNum := rand.Float32()
	intSenderID, _ := strconv.Atoi(senderID)
	probInMat := kv.monkey.matrix[intSenderID][kv.selfID]
	log.Println("Num in Mat is %f", probInMat)
	log.Println("Generated RandNum is %f", randNum)

	return probInMat<randNum //if true message received
}

func (kv *KVService) PutToGetStreamResponse(req *pb.PutRequest, streamHolder pb.KeyValueStore_PutToGetStreamResponseServer) error {
	//TODO
	if kv.monkey != nil{
		pr, ok := peer.FromContext(streamHolder.Context())
		if !ok {
			log.Fatalf("[getClinetIP] invoke FromContext() failed")
		}
		if pr.Addr == net.Addr(nil) {
			log.Fatalf("[getClientIP] peer.Addr is nil")
		}
		senderAddr := pr.Addr.String()
		fmt.Printf(senderAddr)
		if !kv.notToDrop(req.SelfID){
			e := new(PacketLossError)
			e.Msg = "you didnt pass ChaosMonkey"
			return nil
		}
	}
	//TODO
	key, val := req.Key, req.Value
	kv.putLocal(key, val)
	ret := &pb.PutResponse{Ret: pb.ReturnCode_SUCCESS}
	streamHolder.Send(ret)
	return nil
}


func (kv *KVService) parseAndApplyEntry(logEntry entry){
	key, val := logEntry.key, logEntry.val
	kv.dictLock.Lock()
	defer kv.dictLock.Unlock()
	kv.dict[key] = val
}

func (kv *KVService) Start(){
	aPort := "9527"
	lis, err := net.Listen("tcp", ":"+aPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	kv := NewKVService(serverList, addr, selfServerDescriptor.ServerId, monkey)
	pb.RegisterKeyValueStoreServer(grpcServer, kv)
	pb.RegisterRaftServer(grpcServer, kv.raft)
	grpcServer.Serve(lis)

}
