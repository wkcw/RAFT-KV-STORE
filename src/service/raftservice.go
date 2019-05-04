package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	pb "proto"
	"strconv"
	"sync"
	"time"
)

type Membership int

var (
	nameMap = [3]string{"Leader", "Follower", "Candidate"}
)

const (
	Leader    Membership = 0
	Follower  Membership = 1
	Candidate Membership = 2
)

type OutService interface {
	ParseAndApplyEntry(logEntry entry)
}

type RaftService struct {
	state             *State
	membership        Membership
	heartbeatChan     chan bool
	convertToFollower chan bool
	config            *raftConfig
	majorityNum       int
	lastApplied       int64//state lock
	commitIndex       int64//state lock
	appendChan        chan entry
	nextIndex         map[string]int
	matchIndex        map[string]int
	out               OutService
	leaderID          int
	//locks
	rpcMethodLock     *sync.Mutex
	nextIndexLock     *sync.Mutex
	matchIndexLock    *sync.Mutex
	stateLock		  *sync.Mutex
}

func NewRaftService(appendChan chan entry, out OutService) *RaftService {
	state := InitState() //todo
	membership := Follower
	heartbeatChan := make(chan bool)
	convertToFollower := make(chan bool)
	config := createConfig() //todo
	majorityNum := config.ServerList.ServerNum/2 + 1
	commitIndex := int64(-1)
	lastApplied := int64(-1)
	rpcMethodLock := &sync.Mutex{}
	nextIndexLock := &sync.Mutex{}
	matchIndexLock := &sync.Mutex{}
	stateLock := &sync.Mutex{}
	return &RaftService{state: state, membership: membership, heartbeatChan: heartbeatChan,
		convertToFollower: convertToFollower, config: config, majorityNum: majorityNum, commitIndex: commitIndex,
		lastApplied: lastApplied, appendChan: appendChan, out: out, rpcMethodLock: rpcMethodLock,
		nextIndexLock:nextIndexLock, matchIndexLock:matchIndexLock, stateLock:stateLock}
}

func (myRaft *RaftService) AppendEntries(ctx context.Context, req *pb.AERequest) (*pb.AEResponse, error) {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()


	response := &pb.AEResponse{Term: myRaft.state.CurrentTerm}

	//rule 1
	if req.Term < myRaft.state.CurrentTerm {
		response.Success = pb.RaftReturnCode_FAILURE_TERM
		return response, nil
	}

	myRaft.leaderID, _ = strconv.Atoi(req.LeaderId)

	log.Printf("In RPC AE -> Received Heartbeat from %d\n", myRaft.leaderID)
	myRaft.heartbeatChan <- true
	if myRaft.state.CurrentTerm < req.Term {
		myRaft.state.CurrentTerm = req.Term
		myRaft.state.PersistentStore()
		myRaft.convertToFollower <- true
	}

	//rule 2
	if len(myRaft.state.logs.EntryList)-1 < int(req.PrevLogIndex) ||
		(req.PrevLogIndex != -1 && myRaft.state.logs.EntryList[req.PrevLogIndex].term != req.PrevLogTerm){
		response.Success = pb.RaftReturnCode_FAILURE_PREVLOG
		return response, nil
	}
	appendStartIndex := int64(req.PrevLogIndex) + 1
	//rule 3
	for _, reqEntry := range req.Entries {
		if appendStartIndex >= int64(len(myRaft.state.logs.EntryList)){
			break
		}else if myRaft.state.logs.EntryList[appendStartIndex].term != reqEntry.Term {
			myRaft.state.logs.cutEntries(appendStartIndex)
			break
		}
		appendStartIndex++
	}
	//rule 4
	myRaft.state.logs.appendEntries(appendStartIndex, req.Entries)
	myRaft.state.PersistentStore()
	if len(req.Entries)>0{
		log.Printf("IN RPC AE -> Append Entries Done\n")
	}
	//rule 5
	if req.LeaderCommit > myRaft.commitIndex {
		myRaft.commitIndex = minInt64(req.LeaderCommit, int64(len(myRaft.state.logs.EntryList)) - 1)
		for i := myRaft.lastApplied + 1; i <= myRaft.commitIndex; i++ {
			myRaft.out.ParseAndApplyEntry(myRaft.state.logs.EntryList[i])
			myRaft.lastApplied++
		}
	}
	if len(req.Entries)>0{
		log.Printf("IN RPC AE -> Commit And Apply Done\n")
	}
	response.Success = pb.RaftReturnCode_SUCCESS
	return response, nil
}

func (myRaft *RaftService) RequestVote(ctx context.Context, req *pb.RVRequest) (*pb.RVResponse, error) {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()

	if myRaft.state.CurrentTerm < req.Term {
		myRaft.state.CurrentTerm = req.Term
		myRaft.state.PersistentStore()
		myRaft.convertToFollower <- true
		log.Printf("IN RPC RV -> Convert to Follower with HIGH TERM %d from candidate: %s\n", req.Term, req.CandidateID)
	}
	// reply false if term < currentTerm
	if req.Term < myRaft.state.CurrentTerm {
		log.Printf("IN RPC RV -> Rejected RequestVote with LOW TERM %d from candidate: %s\n", req.Term, req.CandidateID)
		return &pb.RVResponse{Term: myRaft.state.CurrentTerm, VoteGranted: false}, nil
	}
	// candidate's log is at least as up-to-date as receiver's log?
	if (myRaft.state.VoteFor == "" || myRaft.state.VoteFor == req.CandidateID) && myRaft.checkMoreUptodate(*req) {
		myRaft.state.VoteFor = req.CandidateID
		myRaft.state.PersistentStore()
		log.Printf("IN RPC RV -> Vote for server %s, term %d", req.CandidateID, req.Term)
		return &pb.RVResponse{Term: req.Term, VoteGranted: true}, nil
	}
	return &pb.RVResponse{Term: req.Term, VoteGranted: false}, nil
}

func (myRaft *RaftService) leaderAppendEntries(isFirstHeartbeat bool) {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()
	for _, server := range myRaft.config.ServerList.Servers {
		if !isFirstHeartbeat && len(myRaft.state.logs.EntryList)-1 >= myRaft.nextIndex[server.Addr] {
			go myRaft.appendEntryToOneFollower(server.Addr)
		}else{
			go myRaft.appendHeartbeatEntryToOneFollower(server.Addr)
		}
	}
}

func (myRaft *RaftService) candidateRequestVotes(winElectionChan chan bool, quit chan bool) {
	countVoteChan := make(chan bool)
	voteCnt := 1
	var quitForRVRoutines []chan bool
	for i, server := range myRaft.config.ServerList.Servers {
		quitForRVRoutines = append(quitForRVRoutines, make(chan bool))
		go myRaft.requestVoteFromOneServer(server.Addr, countVoteChan, quitForRVRoutines[i])
	}
	for {
		select {
		case vote := <-countVoteChan:
			if vote {
				voteCnt++
			}
			if voteCnt >= myRaft.majorityNum {
				winElectionChan <- true
				return
			}
		case <-quit:
			for _, quitForARoutine := range quitForRVRoutines{
				quitForARoutine <- true
			}
			return
		}
	}
	return
}

func (myRaft *RaftService) appendEntriesRoutine(quit chan bool) {
	timeoutTicker := time.NewTicker(time.Duration(myRaft.config.HeartbeatInterval) * time.Millisecond)
	go myRaft.leaderAppendEntries(true)
	for {
		select {
		case <-timeoutTicker.C:
			go myRaft.leaderAppendEntries(false)
		case <-quit:
			timeoutTicker.Stop()
			return
		}
	}
}

func (myRaft *RaftService) randomTimeInterval() time.Duration {
	id, err := strconv.Atoi(myRaft.config.ID)
	if err!=nil{
		log.Fatalf("IN randomTimeInterval %v", err)
	}
	rand.Seed(int64(id))
	upperBound, lowerBound := myRaft.config.ElectionTimeoutUpperBound, myRaft.config.ElectionTimeoutLowerBound
	ret := time.Duration(rand.Int63n(upperBound-lowerBound) + lowerBound)
	return ret * time.Millisecond
}

func (myRaft *RaftService) mainRoutine() {
	for {
		switch myRaft.membership {
		case Leader:
			myRaft.leaderInitVolatileState()
			quit := make(chan bool)
			go myRaft.appendEntriesRoutine(quit)

		Done:
			for {
				select {
				case <-myRaft.convertToFollower:
					myRaft.membership = Follower
					myRaft.state.VoteFor = ""
					myRaft.state.PersistentStore()
					quit <- true
					break Done
				case appendEntry := <-myRaft.appendChan:
					appendEntry.term = myRaft.state.CurrentTerm
					myRaft.state.logs.EntryList = append(myRaft.state.logs.EntryList, appendEntry)
					myRaft.state.PersistentStore()
				}
			}
		case Follower:
			electionTimer := time.NewTimer(myRaft.randomTimeInterval())
			select {
			case <-electionTimer.C:
				myRaft.membership = Candidate
			case <-myRaft.heartbeatChan:
			case <-myRaft.convertToFollower:
				myRaft.membership = Follower
				myRaft.state.VoteFor = ""
				myRaft.state.PersistentStore()
			}
			electionTimer.Stop()
		case Candidate:
			myRaft.state.CurrentTerm++
			log.Printf("Election Start ---- Term:%d\n", myRaft.state.CurrentTerm)
			myRaft.state.VoteFor = myRaft.config.ID
			myRaft.state.PersistentStore()
			electionTimer := time.NewTimer(myRaft.randomTimeInterval())
			winElectionChan := make(chan bool)
			quit := make(chan bool)
			go myRaft.candidateRequestVotes(winElectionChan, quit)
			select {
			case <-electionTimer.C:
				myRaft.membership = Candidate
			case <-winElectionChan:
				myRaft.membership = Leader
			case <-myRaft.convertToFollower:
				myRaft.membership = Follower
				myRaft.state.VoteFor = ""
				myRaft.state.PersistentStore()
			}
			quit <- true
			electionTimer.Stop()

		}
		log.Printf("Membership now: %s\n, term: %d", nameMap[myRaft.membership], myRaft.state.CurrentTerm)
	}
}

func minInt64(num1 int64, num2 int64) int64 {
	if num1 < num2 {
		return num1
	} else {
		return num2
	}
}

func (myRaft *RaftService) appendHeartbeatEntryToOneFollower(serverAddr string) bool {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()
	log.Printf("IN HB -> Send Heartbeat to server: %s\n", serverAddr)
	// Set up a connection to the server.
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	req := &pb.AERequest{Term: myRaft.state.CurrentTerm, LeaderId: myRaft.config.ID, PrevLogIndex: -1, PrevLogTerm: -1,
		Entries: nil, LeaderCommit: myRaft.commitIndex}
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	if e != nil {
		log.Printf("IN HB -> Send HeartbeatEntry to %s failed : %v\n", serverAddr, e)
	} else {
		if ret.Term > myRaft.state.CurrentTerm {
			myRaft.state.CurrentTerm = ret.Term
			myRaft.state.PersistentStore()
			myRaft.convertToFollower <- true
			return false
		} else {
			return true
		}
	}
	return false
}

func (myRaft *RaftService) appendEntryToOneFollower(serverAddr string) {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()
	fmt.Printf("%v\n", myRaft.nextIndex)
	fmt.Printf("%v\n", myRaft.matchIndex)
	log.Printf("IN AE -> Append Entry nextIndex %d to server %s", myRaft.nextIndex[serverAddr], serverAddr)
	prevLogIndex := int64(myRaft.nextIndex[serverAddr] - 1)
	prevLogTerm := int64(-1)
	if prevLogIndex != int64(-1){
		prevLogTerm = myRaft.state.logs.EntryList[prevLogIndex].term
	}
	sendEntry := entryToPbentry(myRaft.state.logs.EntryList[myRaft.nextIndex[serverAddr]])
	sendEntries := make([]*pb.Entry, 1)
	sendEntries[0] = sendEntry
	req := &pb.AERequest{Term: myRaft.state.CurrentTerm, LeaderId: myRaft.config.ID, PrevLogIndex: prevLogIndex,
		PrevLogTerm: prevLogTerm, Entries: sendEntries, LeaderCommit: myRaft.commitIndex}
	// Set up a connection to the server.
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	if e != nil {
		log.Printf("IN AE -> Entry to %s failed RPC error : %v\n", serverAddr, e)
	} else {
		switch ret.Success {
		case pb.RaftReturnCode_SUCCESS:
			log.Printf("IN AE -> Append Entry to %s Succeeded : %v\n", serverAddr, e)
			myRaft.matchIndex[serverAddr] = myRaft.nextIndex[serverAddr]
			myRaft.nextIndex[serverAddr]++
			if int64(myRaft.matchIndex[serverAddr]) > myRaft.commitIndex &&
				myRaft.state.logs.EntryList[myRaft.matchIndex[serverAddr]].term == myRaft.state.CurrentTerm {
				if countGreater(myRaft.matchIndex, myRaft.matchIndex[serverAddr]) >= myRaft.majorityNum {
					myRaft.state.PersistentStore()
					myRaft.commitIndex = int64(myRaft.matchIndex[serverAddr])
					for i := myRaft.lastApplied + 1; i <= myRaft.commitIndex; i++ {
						myRaft.out.ParseAndApplyEntry(myRaft.state.logs.EntryList[i])
						myRaft.lastApplied++
						myRaft.state.logs.EntryList[i].applyChan <- true
						close(myRaft.state.logs.EntryList[i].applyChan)
						myRaft.state.logs.EntryList[i].applyChan = nil
					}
				}
			}
		case pb.RaftReturnCode_FAILURE_TERM:
			myRaft.state.CurrentTerm = ret.Term
			myRaft.state.PersistentStore()
			myRaft.convertToFollower <- true
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			myRaft.nextIndex[serverAddr]--
		}
	}
	return
}

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteChan chan bool, quit chan bool) {
	log.Printf("IN RV -> Send RequestVote to Server: %s\n", serverAddr)

	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	//myRaft.stateLock.Lock()
	lastLogIndex := int64(len(myRaft.state.logs.EntryList) - 1)
	lastLogTerm := int64(-1)
	if lastLogIndex != -1 {
		lastLogTerm = myRaft.state.logs.EntryList[lastLogIndex].term
	}
	req := &pb.RVRequest{Term: myRaft.state.CurrentTerm, CandidateID: myRaft.config.ID,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm}

	//myRaft.stateLock.Unlock()

	defer connManager.gc()
	//var e error


	ret, e := connManager.rpcCaller.RequestVote(connManager.ctx, req)
	if e != nil {
		log.Printf("IN RV -> RPC RV ERROR: %v\n", e)
		return
	}
//Done:
//	// retry requestVote if failed
//	for {
//		select {
//		case <-quit:
//			return
//		default:
//			ret, e = connManager.rpcCaller.RequestVote(connManager.ctx, req)
//			if e == nil {
//				log.Printf("Got RV Request\n")
//				break Done
//			}
//			log.Printf("Retry RV Request to server: %s\n", serverAddr)
//			time.Sleep(100 * time.Millisecond)
//		}
//	}
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()

	if ret.Term > myRaft.state.CurrentTerm {
		myRaft.state.CurrentTerm = ret.Term
		myRaft.state.PersistentStore()
		myRaft.convertToFollower <- true
		log.Printf("IN RV -> Got Higher Term %d from %s, convert to Follower\n", myRaft.state.CurrentTerm, serverAddr)
	} else {
		countVoteChan <- ret.VoteGranted
		log.Printf("IN RV -> Got Vote from %s\n", serverAddr)

	}
	return
}

func (myRaft *RaftService) leaderInitVolatileState() {
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()
	myRaft.nextIndex = make(map[string]int)
	myRaft.matchIndex = make(map[string]int)
	for _, server := range myRaft.config.ServerList.Servers {
		myRaft.nextIndex[server.Addr] = len(myRaft.state.logs.EntryList)
		myRaft.matchIndex[server.Addr] = 0
	}
	return
}

func (myRaft *RaftService) checkMoreUptodate(candidateReq pb.RVRequest) bool {
	localLastLogTerm := int64(-1)
	if len(myRaft.state.logs.EntryList) > 0 {
		localLastLogTerm = myRaft.state.logs.EntryList[len(myRaft.state.logs.EntryList)-1].term
	}
	if candidateReq.LastLogTerm != localLastLogTerm {
		return candidateReq.LastLogTerm > localLastLogTerm
	} else {
		return candidateReq.LastLogIndex >= int64(len(myRaft.state.logs.EntryList)-1)
	}
}

func (myRaft *RaftService) confirmLeadership(confirmationChan chan bool) {
	done := make(chan bool)
	for _, server := range myRaft.config.ServerList.Servers {
		go func() {
			retVal := myRaft.appendHeartbeatEntryToOneFollower(server.Addr)
			log.Printf("ConfirmLeadership AE: retVal =====> %v", retVal)
			done <- retVal
		}()
	}
	ad := 1
	rej := 0
	confirmationTimer := time.NewTimer(time.Duration(500) * time.Millisecond)
	fmt.Printf("Threads are lauched!!!!\n")

Done:
	for {
		select {
		case retVal := <-done:
			log.Printf("Incase <-done: %v, Current Score: ad%d : rej%d", retVal, ad, rej)
			if retVal {
				log.Printf("Inoutterif: %v, Current Score: ad%d : rej%d", retVal, ad, rej)

				ad++
				if ad >= myRaft.majorityNum {
					log.Printf("Inif: %v, Current Score: ad%d : rej%d", retVal, ad, rej)
					confirmationChan <- true
					//close(confirmationChan)
					log.Printf("Will return: %v, Current Score: ad%d : rej%d", retVal, ad, rej)
					return
				}
			} else {
				rej++
				if rej >= myRaft.majorityNum {
					confirmationChan <- false
					//close(confirmationChan)
					log.Printf("Will return: %v, Current Score: ad%d : rej%d", retVal, ad, rej)
					return
				}
			}
			log.Printf("Got retVal: %v, Current Score: ad%d : rej%d", retVal, ad, rej)
		case <-confirmationTimer.C:
			go myRaft.confirmLeadership(confirmationChan)
			log.Printf("ConfirmLeadership Timeout!")
			break Done
		}
	}
}
