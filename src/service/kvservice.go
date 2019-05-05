package service

import (
	pb_monkey "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"net"
	pb "proto"
	"strings"
	"sync"
)



type KVService struct{
	dictLock       *sync.RWMutex
	dict           map[string]string
	selfAddr       string
	raft           *RaftService
	appendChan     chan entry
	clientSequencePairMap     map[string]SequencePair
	SequenceMapLock   *sync.RWMutex
}


func (kv *KVService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error){
	fmt.Printf("############################%v", kv.raft.membership)
	if kv.raft.membership != Leader {
		ret := &pb.GetResponse{Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER, LeaderID: int32(kv.raft.leaderID)}
		return ret, nil
	}
	key := req.Key
	confirmationResultChan := make(chan bool)
	go kv.raft.confirmLeadership(confirmationResultChan)
	iAmLeader := <- confirmationResultChan
	fmt.Printf("confirmation returned\n")
	if iAmLeader{
		data, e := kv.getLocal(key)
		if e == nil {
			ret := &pb.GetResponse{Value: data, Ret: pb.ReturnCode_SUCCESS}
			return ret, nil
		}else{
			ret := &pb.GetResponse{Value: "", Ret: pb.ReturnCode_FAILURE_GET_NOKEY}
			return ret, nil
		}
	}else{
		ret := &pb.GetResponse{Value: "", Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER, LeaderID:-1}
		return ret, nil
	}

}

func (kv *KVService) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error){
	fmt.Printf("Got Put Request\n")
	//if I am not leader, tell client leader ID and Address
	if kv.raft.membership != Leader {
		ret := &pb.PutResponse{Ret: pb.ReturnCode_FAILURE_GET_NOTLEADER, LeaderID: int32(kv.raft.leaderID)}
		return ret, nil
	}
	fmt.Printf("And I think I am Leader?????\n")
	clientIp, err := getClientIP(ctx)
	if err!=nil{
		log.Printf("%v", err)
		ret := &pb.PutResponse{Ret: pb.ReturnCode_FAILURE_PUT_CANTPARSECLIENTIP}
		return ret, nil
	}
	reqSequenceNo := req.SequenceNo
	kv.SequenceMapLock.RLock()
	if recordedSequencePair, ok := kv.clientSequencePairMap[clientIp]; ok!=false{
		if recordedSequencePair.SequenceNo >= reqSequenceNo{
			return &recordedSequencePair.Response, nil
		}
	}

	key, val := req.Key, req.Value
	log.Printf("Dealing Put Request key:%s, val:%s", key, val)
	applyChan := make(chan bool)
	logEntry := entry{op:"put", key:key, val:val, term:-1, applyChan:applyChan}
	kv.appendChan <- logEntry
	log.Printf("Delivered ENtry to Raft logic\n")
	applyStatus := <- applyChan
	log.Printf("Passed <-applyChan log successfullly applied!!\n")
	ret := &pb.PutResponse{}
	if applyStatus{
		kv.putLocal(key, val)
		log.Printf("Successfully applied a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_SUCCESS
		ret.LeaderID = int32(kv.raft.leaderID) // ?
		//kv.clientSequencePairMap[clientIp] = SequencePair{SequenceNo:reqSequenceNo, Response:*ret}
	}else{
		log.Printf("Failed to apply a request with Key: %s, Value: %s", key, val)
		ret.Ret = pb.ReturnCode_FAILURE_PUT
		ret.LeaderID = int32(kv.raft.leaderID) // ?
		//kv.clientSequencePairMap[clientIp] = SequencePair{SequenceNo:reqSequenceNo, Response:*ret}
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



func NewKVService(ID string) *KVService{
	kv := &KVService{dictLock: new(sync.RWMutex), dict:make(map[string]string),
		clientSequencePairMap:make(map[string]SequencePair), SequenceMapLock:new(sync.RWMutex)}
	appendChan := make(chan entry)
	kv.appendChan = appendChan
	raft := NewRaftService(kv.appendChan, kv, ID)
	kv.raft = raft
	return kv
}

func (kv *KVService) ParseAndApplyEntry(logEntry entry){
	key, val := logEntry.key, logEntry.val
	kv.dictLock.Lock()
	defer kv.dictLock.Unlock()
	kv.dict[key] = val
}


func (kv *KVService) Start() {
	aPort := strings.Split(kv.raft.config.SelfAddr, ":")[1]
	lis, err := net.Listen("tcp", ":" + aPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	go kv.raft.mainRoutine()
	pb.RegisterKeyValueStoreServer(grpcServer, kv)
	pb.RegisterRaftServer(grpcServer, kv.raft)
	pb_monkey.RegisterChaosMonkeyServer(grpcServer, kv.raft.monkey)
	grpcServer.Serve(lis)
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}