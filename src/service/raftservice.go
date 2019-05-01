package service

import (
	pb "proto"
	"context"
	"time"
	"log"
	"sync"
	"strconv"
)

type Membership int

const (

	Leader      Membership = 0
	Follower      Membership = 1
	Candidate      Membership = 2
)

type OutService interface {
	ParseAndApplyEntry(logEntry entry)
}

type RaftService struct{
	state *State
	membership Membership
	heartbeatChan chan bool
	convertToFollower chan bool
	config *raftConfig
	majorityNum int
	lastApplied int64
	commitIndex int64
	appendChan chan entry
	rpcMethodLock *sync.Mutex
	nextIndex map[string]int
	matchIndex map[string]int
	out OutService
	leaderID int
}

func NewRaftService(appendChan chan entry, out OutService) *RaftService{
	state := InitState() //todo
	membership := Follower
	heartbeatChan := make(chan bool)
	convertToFollower := make(chan bool)
	config := createConfig()//todo
	majorityNum := config.serverList.serverNum / 2 +1
	commitIndex := int64(-1)
	rpcMethodLock := &sync.Mutex{}
	return &RaftService{state:state, membership:membership, heartbeatChan:heartbeatChan,
						convertToFollower:convertToFollower, config: config, majorityNum:majorityNum, commitIndex:commitIndex,
						appendChan:appendChan, rpcMethodLock:rpcMethodLock, out:out}
}


func (myRaft *RaftService) AppendEntries(ctx context.Context, req *pb.AERequest) (*pb.AEResponse, error){
	myRaft.rpcMethodLock.Lock()
	defer myRaft.rpcMethodLock.Unlock()

	response := &pb.AEResponse{Term:myRaft.state.CurrentTerm}

	//rule 1
	if req.Term < myRaft.state.CurrentTerm {
		response.Success = pb.RaftReturnCode_FAILURE_TERM
		return response, nil
	}

	myRaft.leaderID, _ = strconv.Atoi(req.LeaderId)

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
			myRaft.state.logs.EntryList[i].applyChan <- true
			close(myRaft.state.logs.EntryList[i].applyChan)
			myRaft.state.logs.EntryList[i].applyChan = nil
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

func (myRaft *RaftService) leaderAppendEntries(isFirstHeartbeat bool){
	for _, server := range myRaft.config.serverList.servers{
		if !isFirstHeartbeat && len(myRaft.state.logs.EntryList)-1 >= myRaft.nextIndex[server.addr]{
			go myRaft.appendEntryToOneFollower(server.addr)
		}
		go myRaft.appendHeartbeatEntryToOneFollower(server.addr)
	}
}

func (myRaft *RaftService) candidateRequestVotes(winElectionChan chan bool, quit chan bool){
	countVoteChan := make(chan bool)
	voteCnt := 0
	for _, server := range myRaft.config.serverList.servers{
		go myRaft.requestVoteFromOneServer(server.addr, countVoteChan, quit)
	}
	for {
		select {
		case vote := <-countVoteChan:
			if vote {
				voteCnt++
			}
			if voteCnt > myRaft.majorityNum {
				winElectionChan <- true
				return
			}
		case <-quit:
			return
		}
	}
	return
}

func (myRaft *RaftService) appendEntriesRoutine(quit chan bool){
	timeoutTicker := time.NewTicker(time.Duration(myRaft.config.heartbeatInterval) * time.Millisecond)
	go myRaft.leaderAppendEntries(true)
	for {
		select {
		case <- timeoutTicker.C:
			go myRaft.leaderAppendEntries(false)
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
				for{
					select{
					case <- myRaft.convertToFollower:
						myRaft.membership = Follower
						myRaft.state.VoteFor = ""
						quit <- true
						break
					case appendEntry := <- myRaft.appendChan:
						myRaft.state.logs.EntryList = append(myRaft.state.logs.EntryList, appendEntry)
					}
				}
			case Follower:
				electionTimer := time.NewTimer(myRaft.randomTimeInterval())
				select {
				case <- electionTimer.C:
					myRaft.membership = Candidate
				case <- myRaft.heartbeatChan:
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
				}
				electionTimer.Stop()
			case Candidate:
				myRaft.state.CurrentTerm++;
				myRaft.state.VoteFor = myRaft.config.ID
				electionTimer := time.NewTimer(myRaft.randomTimeInterval())
				winElectionChan := make(chan bool)
				quit := make(chan bool)
				go myRaft.candidateRequestVotes(winElectionChan, quit)
				select {
				case <- electionTimer.C:
					myRaft.membership = Candidate
				case <- winElectionChan:
					myRaft.membership = Leader
				case <- myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
					quit <- true
				}
				electionTimer.Stop()

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

func (myRaft *RaftService)appendHeartbeatEntryToOneFollower(serverAddr string) bool{
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
			return false
		}else{
			return true
		}
	}
	return false
}

func (myRaft *RaftService)appendEntryToOneFollower(serverAddr string){
	prevLogIndex := int64(myRaft.nextIndex[serverAddr]-1)
	prevLogTerm := myRaft.state.logs.EntryList[prevLogIndex].term
	sendEntry := entryToPbentry(myRaft.state.logs.EntryList[myRaft.nextIndex[serverAddr]])
	sendEntries := make([]*pb.Entry, 1)
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
		switch ret.Success{
		case pb.RaftReturnCode_SUCCESS:
			myRaft.matchIndex[serverAddr] = myRaft.nextIndex[serverAddr]
			myRaft.nextIndex[serverAddr]++
			if int64(myRaft.matchIndex[serverAddr]) > myRaft.commitIndex &&
				myRaft.state.logs.EntryList[myRaft.matchIndex[serverAddr]].term == myRaft.state.CurrentTerm {
				if countGreater(myRaft.matchIndex, myRaft.matchIndex[serverAddr]) >= myRaft.majorityNum{
					myRaft.commitIndex = int64(myRaft.matchIndex[serverAddr])
					for i:=myRaft.lastApplied+1; i<=myRaft.commitIndex; i++{
						myRaft.out.ParseAndApplyEntry(myRaft.state.logs.EntryList[i])
						myRaft.state.logs.EntryList[i].applyChan <- true
						close(myRaft.state.logs.EntryList[i].applyChan)
						myRaft.state.logs.EntryList[i].applyChan = nil
					}
				}
			}
		case pb.RaftReturnCode_FAILURE_TERM:
			myRaft.state.CurrentTerm = ret.Term
			myRaft.convertToFollower <- true
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			myRaft.nextIndex[serverAddr]--
		}
	}
	return
}

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteChan chan bool, quit chan bool){
	connManager := createConnManager(serverAddr)
	req := &pb.RVRequest{Term:myRaft.state.CurrentTerm, CandidateID:myRaft.config.ID,
							LastLogIndex:myRaft.state.logs.lastIndex,
							LastLogTerm:myRaft.state.logs.EntryList[myRaft.state.logs.lastIndex].term}
	defer connManager.gc()
	var ret *pb.RVResponse
	var e error

	// retry requestVote if failed
	for {
		select{
		case <-quit:
			return
		default:
			ret, e = connManager.rpcCaller.RequestVote(connManager.ctx, req)
			if e==nil {
				break
			}
		}
	}
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
	for _, server := range myRaft.config.serverList.servers{
		myRaft.nextIndex[server.addr] = len(myRaft.state.logs.EntryList)
		myRaft.matchIndex[server.addr] = 0
	}
	return
}

func (myRaft * RaftService)confirmLeadership(confirmationChan chan bool){
	done := make(chan bool)
	for _, server := range myRaft.config.serverList.servers{
		go func(){
			retVal := myRaft.appendHeartbeatEntryToOneFollower(server.addr)
			done <- retVal
		}()
	}
	ad := 1
	rej := 0
	confirmationTimer := time.NewTimer(time.Duration(10) * time.Millisecond)
	for{
		select{
		case retVal := <-done:
			if retVal{
				ad++
				if ad >= myRaft.config.serverList.serverNum / 2 + 1{
					confirmationChan <- true
					close(confirmationChan)
					return
				}
			}else{
				rej++
				if rej >= myRaft.config.serverList.serverNum / 2 + 1{
					confirmationChan <- false
					close(confirmationChan)
					return
				}
			}
		case <- confirmationTimer.C:
			go myRaft.confirmLeadership(confirmationChan)
			break
		}
	}
}