package service

import (
	pb "proto"
	"context"
	"time"
	"log"
	"sync"
)

type Membership int

const (

	Leader      Membership = 0
	Follower      Membership = 1
	Candidate      Membership = 2
)

type RaftService struct{
	state *State
	membership Membership
	heartbeatChan chan bool
	convertToFollower chan bool
	config raftConfig
	commitIndex int64
	dict *map[string]string
	applyChan chan entry
	rpcMethodLock sync.Mutex
	nextIndex map[string]int
	matchIndex map[string]int
}

func NewRaftService() *RaftService{
	return &RaftService{}
}


func (myRaft *RaftService) AppendEntries(ctx context.Context, req pb.AERequest) (*pb.AEResponse, error){
	myRaft.rpcMethodLock.Lock()
	defer myRaft.rpcMethodLock.Unlock()

	response := &pb.AEResponse{Term:myRaft.state.CurrentTerm}

	//rule 1
	if req.Term < myRaft.state.CurrentTerm {
		response.Success = pb.RaftReturnCode_FAILURE_TERM
		return response, nil
	}

	myRaft.heartbeatChan <- true
	if myRaft.state.CurrentTerm < req.Term{
		myRaft.state.CurrentTerm = req.Term
		myRaft.convertToFollower <- true
	}
	//rule 2
	if len(myRaft.state.logs.EntryList)-1 < int(req.PrevLogIndex) || myRaft.state.logs.EntryList[req.PrevLogIndex].term != req.PrevLogTerm{
		response.Success = pb.RaftReturnCode_FAILURE_PREVLOG
		return response, nil

	}
	appendStartIndex := int64(req.PrevLogIndex)+1
	//rule 3
	for _, reqEntry := range req.Entries{
		if myRaft.state.logs.EntryList[appendStartIndex].term != reqEntry.Term {
			myRaft.state.logs.cutEntries(appendStartIndex)
			break
		}
		appendStartIndex++
	}
	//rule 4
	myRaft.state.logs.appendEntries(appendStartIndex, req.Entries)
	//rule 5
	if req.LeaderCommit > myRaft.state.logs.commitIndex {
		tmpCommitIndex := minInt64(req.LeaderCommit, int64(len(myRaft.state.logs.EntryList)))
		for i:= myRaft.state.logs.commitIndex; i<tmpCommitIndex; i++{
			myRaft.applyChan <- myRaft.state.logs.EntryList[i]
		}
		myRaft.state.logs.commitIndex =  minInt64(req.LeaderCommit, int64(len(myRaft.state.logs.EntryList)))
	}
	response.Success = pb.RaftReturnCode_SUCCESS
	return response, nil
}

func (myRaft *RaftService) RequestVote(ctx context.Context, req *pb.RVRequest) (*pb.RVResponse, error){
	myRaft.rpcMethodLock.Lock()
	defer myRaft.rpcMethodLock.Unlock()

	if myRaft.state.CurrentTerm < req.Term{
		myRaft.state.CurrentTerm = req.Term
		myRaft.convertToFollower <- true
	}
	// reply false if term < currentTerm
	if req.Term < myRaft.state.CurrentTerm {
		return &pb.RVResponse{Term: myRaft.state.CurrentTerm, VoteGranted: false}, nil
	}
	// candidate's log is at least as up-to-date as receiver's log?
	if myRaft.state.VoteFor == "" || myRaft.state.VoteFor == req.CandidateID {
		myRaft.state.VoteFor = req.CandidateID
		return &pb.RVResponse{Term: req.Term, VoteGranted: true}, nil
	}
	return &pb.RVResponse{Term: req.Term, VoteGranted: false}, nil
}

func (myRaft *RaftService) leaderAppendEntries(){
	for _, serverAddr := range myRaft.config.peers{
		if len(myRaft.state.logs.EntryList)-1 >= myRaft.nextIndex[serverAddr]{
			go myRaft.appendEntryToOneFollower(serverAddr)
		}
		go myRaft.appendHeartbeatEntryToOneFollower(serverAddr)
	}
}

func (myRaft *RaftService) candidateRequestVotes(winElectionChan chan bool){
	countVoteChan := make(chan bool)
	voteCnt := 0
	for _, serverAddr := range myRaft.config.peers{
		go myRaft.requestVoteFromOneServer(serverAddr, countVoteChan)
	}
	for vote := range countVoteChan{
		if vote{
			voteCnt++
		}
		if voteCnt > myRaft.config.majorityNum{
			winElectionChan <- true
			break
		}
	}
	return
}

func (myRaft *RaftService) appendEntriesRoutine(quit chan bool){
	timeoutTicker := time.NewTicker(time.Duration(myRaft.config.heartbeatInterval) * time.Millisecond)
	go myRaft.leaderAppendEntries()
	for {
		select {
		case <- timeoutTicker.C:
			go myRaft.leaderAppendEntries()
		case <- quit:
			timeoutTicker.Stop()
			return
		}
	}
}

func (myRaft * RaftService) randomTimeInterval() time.Duration{
	return 50 * time.Millisecond
}

func (myRaft *RaftService) mainRoutine(){
	for {
		switch myRaft.membership {
			case Leader:
				myRaft.leaderInitVolatileState()
				quit := make(chan bool)
				go myRaft.appendEntriesRoutine(quit)
				select{
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
					quit <- true
				}
			case Follower:
				electionTimer := time.NewTimer(myRaft.randomTimeInterval())
				select {
				case <- electionTimer.C:
					myRaft.membership = Candidate
				case <- myRaft.heartbeatChan:
					electionTimer.Stop()
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
				}
			case Candidate:
				myRaft.state.CurrentTerm++;
				myRaft.state.VoteFor = myRaft.config.ID
				electionTimer := time.NewTimer(myRaft.randomTimeInterval())
				winElectionChan := make(chan bool)
				go myRaft.candidateRequestVotes(winElectionChan)
				select {
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
				case <- electionTimer.C:
					myRaft.membership = Candidate
				case <- winElectionChan:
					electionTimer.Stop()
					myRaft.membership = Leader
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
				}

		}
	}
}

func minInt64(num1 int64, num2 int64) int64{
	if num1<num2 {
		return num1
	}else{
		return num2
	}
}

func (myRaft *RaftService)appendHeartbeatEntryToOneFollower(serverAddr string){
	// Set up a connection to the server.
	connManager := createConnManager(serverAddr)
	req := &pb.AERequest{Term:myRaft.state.CurrentTerm, LeaderId:myRaft.config.ID, PrevLogIndex:-1, PrevLogTerm:-1,
		Entries:nil, LeaderCommit:myRaft.commitIndex}
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	if e!=nil{
		log.Printf("Send HeartbeatEntry to %s failed : %v\n", serverAddr, e)
	}else{
		if ret.Term>myRaft.state.CurrentTerm{
			myRaft.state.CurrentTerm = ret.Term
			myRaft.convertToFollower <- true
		}
	}
	return 
}

func (myRaft *RaftService)appendEntryToOneFollower(serverAddr string){
	prevLogIndex := int64(myRaft.nextIndex[serverAddr]-1)
	prevLogTerm := myRaft.state.logs.EntryList[prevLogIndex].term
	sendEntry := entryToPbentry(myRaft.state.logs.EntryList[myRaft.nextIndex[serverAddr]])
	sendEntries := make([]*pb.AERequest_Entry, 1)
	sendEntries[0] = sendEntry
	req := &pb.AERequest{Term:myRaft.state.CurrentTerm, LeaderId:myRaft.config.ID, PrevLogIndex:prevLogIndex,
						PrevLogTerm:prevLogTerm, Entries:sendEntries, LeaderCommit:myRaft.commitIndex}
	// Set up a connection to the server.
	connManager := createConnManager(serverAddr)
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	if e!=nil{
		log.Printf("Send HeartbeatEntry to %s failed : %v\n", serverAddr, e)
	}else{
		if ret.Term>myRaft.state.CurrentTerm{
			myRaft.state.CurrentTerm = ret.Term
			myRaft.convertToFollower <- true
		}
	}
	return
}

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteChan chan bool){
	connManager := createConnManager(serverAddr)
	req := &pb.RVRequest{Term:myRaft.state.CurrentTerm, CandidateID:myRaft.config.ID,
							LastLogIndex:myRaft.state.logs.lastIndex,
							LastLogTerm:myRaft.state.logs.EntryList[myRaft.state.logs.lastIndex].term}
	defer connManager.gc()
	var ret *pb.RVResponse
	var e error

	// retry requestVote if failed
	requestVoteRetryCnt := 0
	for {
		requestVoteRetryCnt++
		ret, e = connManager.rpcCaller.RequestVote(connManager.ctx, req)
		if e==nil || requestVoteRetryCnt==myRaft.config.requestVoteRetryNum{
			break
		}
	}
	//if e!=nil{
	//	log.Printf("Send RequestVote to %s failed : %v\n", serverAddr, e)
	//
	//}else{
	//	if ret.Term>myRaft.state.CurrentTerm{
	//		myRaft.state.CurrentTerm = ret.Term
	//		myRaft.convertToFollower <- true
	//	}else{
	//		countVoteChan <- ret.VoteGranted
	//	}
	//}
	if ret.Term>myRaft.state.CurrentTerm{
		myRaft.state.CurrentTerm = ret.Term
		myRaft.convertToFollower <- true
	}else{
		countVoteChan <- ret.VoteGranted
	}
	return
}

func (myRaft *RaftService)leaderInitVolatileState(){
	myRaft.nextIndex = make(map[string]int)
	myRaft.matchIndex = make(map[string]int)
	for _, serverAddr := range myRaft.config.peers{
		myRaft.nextIndex[serverAddr] = len(myRaft.state.logs.EntryList)
		myRaft.matchIndex[serverAddr] = 0
	}
	return
}
