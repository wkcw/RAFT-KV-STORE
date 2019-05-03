package service

import (
	pb_monkey "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "proto"
	"sync"
	"strings"
)



type KVService struct{
	dictLock       *sync.RWMutex
	dict           map[string]string
	selfAddr       string
	raft           *RaftService
	appendChan     chan entry
	//monkey         *MonkeyService
	//addrToID       map[string]int
	//idToAddr       map[int]string
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

func (kv *KVService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error){
	if kv.raft.membership != Leader {
		ret := &pb.GetResponse{Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER, LeaderID: int32(kv.raft.leaderID)}
		return ret, nil
	}
	key := req.Key
	confirmationResultChan := make(chan bool)
	kv.raft.confirmLeadership(confirmationResultChan)
	iAmLeader := <- confirmationResultChan
	if iAmLeader{
		data, e := kv.getLocal(key)
		if e == nil {
			ret := &pb.GetResponse{Value: data, Ret: pb.ReturnCode_SUCCESS}
			return ret, nil
		}else{
			ret := &pb.GetResponse{Value: "", Ret: pb.ReturnCode_FAILURE_GET_NOKEY}
			return ret, e
		}
	}else{
		ret := &pb.GetResponse{Value: "", Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER}
		return ret, nil
	}

}

func (kv *KVService) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	//if I am not leader, tell client leader ID and Address
	if kv.raft.membership != Leader {
		ret := &pb.PutResponse{Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER, LeaderID: int32(kv.raft.leaderID)}
		return ret, nil
	}
	key, val := req.Key, req.Value
	log.Printf("Dealing Put Request key:%s, val:%s", key, val)
	applyChan := make(chan bool)
	logEntry := entry{op:"put", key:key, val:val, term:-1, applyChan:applyChan}
	kv.appendChan <- logEntry
	applyStatus := <- applyChan
	ret := &pb.PutResponse{}
	if applyStatus{
		kv.putLocal(key, val)
		log.Printf("Successfully applied a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_SUCCESS
		ret.LeaderID = int32(kv.raft.leaderID) // ?
	}else{
		log.Printf("Failed to apply a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_FAILURE_PUT
		ret.LeaderID = int32(kv.raft.leaderID) // ?
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



func NewKVService() *KVService{
	kv := &KVService{dictLock: new(sync.RWMutex), dict:make(map[string]string)}
	appendChan := make(chan entry)
	kv.appendChan = appendChan
	raft := NewRaftService(kv.appendChan, kv)
	kv.raft = raft
	return kv
}

func NewMonkeyService(n int) *MonkeyService {
	return &MonkeyService{matrix: make([][]float32, n, n)}
}

//func (kv *KVService) notToDrop(senderID string) bool{
//	log.Printf("senderID is %s, self ID is %d\n", senderID, kv.selfID)
//	randNum := rand.Float32()
//	intSenderID, _ := strconv.Atoi(senderID)
//	probInMat := kv.monkey.matrix[intSenderID][kv.selfID]
//	log.Println("Num in Mat is %f", probInMat)
//	log.Println("Generated RandNum is %f", randNum)
//
//	return probInMat<randNum //if true message received
//}


func (kv *KVService) ParseAndApplyEntry(logEntry entry){
	key, val := logEntry.key, logEntry.val
	kv.dictLock.Lock()
	defer kv.dictLock.Unlock()
	kv.dict[key] = val
}


func (kv *KVService) Start() {
	aPort := strings.Split(kv.raft.config.SelfAddr, ":")[1]
	lis, err := net.Listen("tcp", ":"+aPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	go kv.raft.mainRoutine()
	pb.RegisterKeyValueStoreServer(grpcServer, kv)
	pb.RegisterRaftServer(grpcServer, kv.raft)
	grpcServer.Serve(lis)
}
