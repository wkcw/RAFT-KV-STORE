package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	pb "proto"
	"strconv"
	"sync"
	"time"
)

type Membership int

var (
	nameMap = [3]string{"Leader", "Follower", "Candidate"}
	uselock = true
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
	exitChan   	      chan bool
	state             *State
	membership        Membership
	heartbeatChan     chan bool
	grantVoteChan chan bool
	leaderToFollowerChan chan bool
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
	rpcMethodLock     *sync.RWMutex
	nextIndexLock     *sync.Mutex
	matchIndexLock    *sync.Mutex
	stateLock		  *sync.RWMutex
	monkey 		      *MonkeyService
}

func NewRaftService(appendChan chan entry, out OutService, ID string) *RaftService {
	exitChan := make(chan bool)
	membership := Follower
	heartbeatChan := make(chan bool, 100)
	grantVoteChan := make(chan bool, 100)
	leaderToFollowerChan := make(chan bool, 1)
	config := createConfig(ID) //todo
	majorityNum := config.ServerList.ServerNum/2 + 1
	commitIndex := int64(-1)
	lastApplied := int64(-1)
	rpcMethodLock := new(sync.RWMutex)
	nextIndexLock := &sync.Mutex{}
	matchIndexLock := &sync.Mutex{}
	stateLock := new(sync.RWMutex)
	selfID, _ := strconv.ParseInt(config.ID, 10, 32)
	state := InitState(selfID)
	monkey := NewMonkeyService(config.ServerList.ServerNum, int32(selfID))
	return &RaftService{exitChan: exitChan, state: state, membership: membership, heartbeatChan: heartbeatChan,
		grantVoteChan: grantVoteChan, leaderToFollowerChan:leaderToFollowerChan, config: config, majorityNum: majorityNum, commitIndex: commitIndex,
		lastApplied: lastApplied, appendChan: appendChan, out: out, rpcMethodLock: rpcMethodLock,
		nextIndexLock:nextIndexLock, matchIndexLock:matchIndexLock, stateLock:stateLock, monkey: monkey}
}
func (myRaft *RaftService) dropMessage (senderID int64) bool{
	selfID, _:= strconv.ParseInt(myRaft.config.ID, 10, 32)
	//log.Printf("senderID is %d, self ID is %d\n", senderID, selfID)
	randNum := rand.Float32()
	probInMat := myRaft.monkey.matrix[senderID][selfID]
	//log.Printf("Num in Mat is %f\n", probInMat)
	//log.Printf("Generated RandNum is %f\n", randNum)
	return randNum < probInMat
}
func (myRaft *RaftService) AppendEntries(ctx context.Context, req *pb.AERequest) (*pb.AEResponse, error) {


	//TODO add drop message
	if myRaft.dropMessage(req.Sender) {
		log.Println("Dropping the message in AppendEntries")
		time.Sleep(time.Duration(10 * time.Second))
		ret := &pb.AEResponse{Term: req.Term, Success: pb.RaftReturnCode_SUCCESS}
		return ret, nil;
	}

	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("RPC AppendEntries acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("RPC AppendEntries released lock\n")}()
	}
	//log.Printf("In RPC AE -> Successfully acquired lock\n")




	response := &pb.AEResponse{Term: myRaft.state.CurrentTerm}

	//rule 1
	if req.Term < myRaft.state.CurrentTerm {
		response.Success = pb.RaftReturnCode_FAILURE_TERM
		return response, nil
	}

	myRaft.leaderID, _ = strconv.Atoi(req.LeaderId)

	log.Printf("In RPC AE -> Received Heartbeat from %d\n", myRaft.leaderID)




	log.Printf("IN RPC AE -> Before send true to heartbeatChan")

	myRaft.heartbeatChan <- true
	if myRaft.state.CurrentTerm < req.Term {
		log.Printf("IN RPC AE -> Before convertToFollower")
		myRaft.changeToFollower(req.Term)
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
	//TODO add drop message
	if myRaft.dropMessage(req.Sender) {
		log.Println("Dropping the message in RequestVote")
		time.Sleep(time.Duration(10 * time.Second))
		ret := &pb.RVResponse{Term: req.Term, VoteGranted: false}
		return ret, nil;
	}
	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("RPC RequestVote acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("RPC RequestVote released lock\n")}()
	}
	//log.Printf("In RPC RV -> Successfully acquired lock\n")

	if myRaft.state.CurrentTerm < req.Term {
		log.Printf("IN RPC RV -> Before convertToFollower")
		myRaft.changeToFollower(req.Term)
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

func (myRaft *RaftService) leaderAppendEntries(isFirstHeartbeat bool, reqTerm int64) {
	if uselock{
		myRaft.stateLock.RLock()
		log.Printf("leaderAppendEntries acquired lock\n")
		defer func(){myRaft.stateLock.RUnlock(); log.Printf("leaderAppendEntries released lock\n")}()
	}
	for _, server := range myRaft.config.ServerList.Servers {
		log.Printf("nextIndex of Server %s is %d, myLogLen-1 is %d", server.Addr, myRaft.nextIndex[server.Addr], len(myRaft.state.logs.EntryList)-1)
		if !isFirstHeartbeat && len(myRaft.state.logs.EntryList)-1 >= myRaft.nextIndex[server.Addr] {
			go myRaft.appendEntryToOneFollower(server.Addr, reqTerm)
		}else{
			go myRaft.appendHeartbeatEntryToOneFollower(server.Addr, reqTerm)
		}
	}
}

func (myRaft *RaftService) candidateRequestVotes(winElectionChan chan bool, quit chan bool, reqTerm int64) {
	countVoteChan := make(chan bool, len(myRaft.config.ServerList.Servers))
	voteCnt := 1
	for _, server := range myRaft.config.ServerList.Servers {
		go myRaft.requestVoteFromOneServer(server.Addr, countVoteChan, reqTerm)
	}
	voteNum := 0
	for {
		select {
		case vote := <-countVoteChan:
			voteNum++
			log.Printf("Collected %d Vote\n", voteNum)
			if vote {
				voteCnt++
			}
			if voteCnt >= myRaft.majorityNum {
				winElectionChan <- true
				//close(countVoteChan)
				log.Printf("Won Election!!!\n")
				return
			}
		case <-quit:
			//for _, quitForARoutine := range quitForRVRoutines{
			//	quitForARoutine <- true
			//}
			return
		}
	}
	return
}

func (myRaft *RaftService) appendEntriesRoutine(reqTerm int64) {
	timeoutTicker := time.NewTicker(time.Duration(myRaft.config.HeartbeatInterval) * time.Millisecond)
	go myRaft.leaderAppendEntries(true, reqTerm)
	for {
		select {
		case <-timeoutTicker.C:
			if myRaft.state.CurrentTerm > reqTerm{
				return
			}
			go myRaft.leaderAppendEntries(false, reqTerm)
		//case <-quit:

		}
	}
}



func (myRaft *RaftService) randomTimeInterval() time.Duration {
	rand.Seed(time.Now().Unix())
	upperBound, lowerBound := myRaft.config.ElectionTimeoutUpperBound, myRaft.config.ElectionTimeoutLowerBound
	ret := time.Duration(rand.Int63n(upperBound-lowerBound) + lowerBound)
	return ret * time.Millisecond
}

func (myRaft *RaftService) mainRoutine() {
	for {
		switch myRaft.membership {
		case Leader:
			select {
			case <- myRaft.exitChan:
				fmt.Println("Exit from leader state!")
				os.Exit(1)
			case appendEntry := <-myRaft.appendChan:
				if uselock{
					myRaft.stateLock.Lock()
				}
				appendEntry.term = myRaft.state.CurrentTerm
				myRaft.state.logs.EntryList = append(myRaft.state.logs.EntryList, appendEntry)
				myRaft.state.PersistentStore()
				if uselock{
					myRaft.stateLock.Unlock()
				}
			case <-myRaft.leaderToFollowerChan:

			}

		case Follower:
			electionTimer := time.NewTimer(myRaft.randomTimeInterval())
			select {
			//case <- myRaft.exitChan:
			//	return
			case <-electionTimer.C:
				myRaft.membership = Candidate
			case <-myRaft.heartbeatChan:
			case <- myRaft.grantVoteChan:
			}
			electionTimer.Stop()
		case Candidate:
			if uselock{
				myRaft.stateLock.Lock()
			}
			myRaft.state.CurrentTerm++
			log.Printf("Election Start ---- Term:%d\n", myRaft.state.CurrentTerm)
			myRaft.state.VoteFor = myRaft.config.ID
			myRaft.state.PersistentStore()
			if uselock{
				myRaft.stateLock.Unlock()
			}
			electionTimer := time.NewTimer(myRaft.randomTimeInterval())
			winElectionChan := make(chan bool)
			quit := make(chan bool)
			go myRaft.candidateRequestVotes(winElectionChan, quit, myRaft.state.CurrentTerm)
			select {
			case <- myRaft.exitChan:
				return
			case <-electionTimer.C:
				quit <- true
			case <-winElectionChan:
				myRaft.membership = Leader
				myRaft.leaderInitVolatileState()
				go myRaft.appendEntriesRoutine(myRaft.state.CurrentTerm)
			case <-myRaft.heartbeatChan:
				myRaft.membership = Follower
				quit <- true
			}
			electionTimer.Stop()

		}
		log.Printf("\nMembership now: %s, term: %d, voteFor: %s\n\n", nameMap[myRaft.membership],
			myRaft.state.CurrentTerm, myRaft.state.VoteFor)
	}
}

func minInt64(num1 int64, num2 int64) int64 {
	if num1 < num2 {
		return num1
	} else {
		return num2
	}
}

func (myRaft *RaftService) appendHeartbeatEntryToOneFollower(serverAddr string, reqTerm int64) bool {
	log.Printf("IN HB -> Send Heartbeat to server: %s\n", serverAddr)

	if uselock{
		myRaft.stateLock.RLock()
		log.Printf("appendHeartbeatEntryToOneFollower acquired Rlock\n")
	}
	senderId, convErr := strconv.Atoi(myRaft.config.ID)
	if convErr != nil{
		log.Printf("cant convert ID\n")
		return false
	}
	//if uselock{
	//	myRaft.stateLock.RLock()
	//}
	prevLogTerm := int64(-1)
	if myRaft.nextIndex[serverAddr]>0{
		prevLogTerm = myRaft.state.logs.EntryList[myRaft.nextIndex[serverAddr]-1].term
	}
	req := &pb.AERequest{Term: myRaft.state.CurrentTerm, LeaderId: myRaft.config.ID,
		PrevLogIndex: int64(myRaft.nextIndex[serverAddr]-1), PrevLogTerm: prevLogTerm,
		Entries: nil, LeaderCommit: myRaft.commitIndex, Sender:int64(senderId)}
	if uselock{
		myRaft.stateLock.RUnlock()
		log.Printf("appendHeartbeatEntryToOneFollower released Rlock\n")
	}

	// Set up a connection to the server.
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()


	if e != nil {
		log.Printf("IN HB -> Send HeartbeatEntry to %s failed : %v\n", serverAddr, e)
	} else {
		if uselock{
			myRaft.stateLock.Lock()
			log.Printf("appendHeartbeatEntryToOneFollower acquired lock\n")
			defer func(){myRaft.stateLock.Unlock(); log.Printf("appendHeartbeatEntryToOneFollower released lock\n")}()
		}
		if reqTerm != myRaft.state.CurrentTerm{
			return false
		}
		switch ret.Success {
		case pb.RaftReturnCode_SUCCESS:
			log.Printf("IN HB -> heartbeat to %s PREVLOG Success \n", serverAddr)
			myRaft.matchIndex[serverAddr] = myRaft.nextIndex[serverAddr] - 1
			return true
		case pb.RaftReturnCode_FAILURE_TERM:
			log.Printf("IN HB -> heartbeat to %s TERM FAILURE \n", serverAddr)
			myRaft.state.PersistentStore()
			log.Printf("IN HB -> Before convertToFollower")
			myRaft.changeToFollower(ret.Term)
			dropAndSet(myRaft.leaderToFollowerChan)
			return false
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			log.Printf("IN HB -> heartbeat to %s PREVLOG FAILURE \n", serverAddr)
			myRaft.nextIndex[serverAddr]--
			return false
		}
	}
	return false
}

func (myRaft *RaftService) appendEntryToOneFollower(serverAddr string, reqTerm int64) {
	if uselock{
		myRaft.stateLock.RLock()
		log.Printf("appendEntryToOneFollower acquired Rlock\n")
	}

	log.Printf("IN AE -> Append Entry nextIndex %d to server %s", myRaft.nextIndex[serverAddr], serverAddr)
	prevLogIndex := int64(myRaft.nextIndex[serverAddr] - 1)
	prevLogTerm := int64(-1)
	if prevLogIndex != int64(-1){
		prevLogTerm = myRaft.state.logs.EntryList[prevLogIndex].term
	}

	sendEntries := make([]*pb.Entry, 0)
	for i:= myRaft.nextIndex[serverAddr]; i<len(myRaft.state.logs.EntryList); i++{
		sendEntries = append(sendEntries, entryToPbentry(myRaft.state.logs.EntryList[i]))
	}

	senderId, convErr := strconv.Atoi(myRaft.config.ID)
	if convErr != nil{
		log.Printf("cant convert ID\n")
		//myRaft.stateLock.RUnlock()
		return
	}

	req := &pb.AERequest{Term: myRaft.state.CurrentTerm, LeaderId: myRaft.config.ID, PrevLogIndex: prevLogIndex,
		PrevLogTerm: prevLogTerm, Entries: sendEntries, LeaderCommit: myRaft.commitIndex, Sender:int64(senderId)}
	if uselock{
		myRaft.stateLock.RUnlock();
		log.Printf("appendEntryToOneFollower released Rlock\n")
	}

	// Set up a connection to the server.
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()


	if e != nil {
		log.Printf("IN AE -> Entry to %s failed RPC error : %v\n", serverAddr, e)
	} else {
		if uselock{
			myRaft.stateLock.Lock()
			log.Printf("appendEntryToOneFollower acquired Rlock\n")
			defer func(){myRaft.stateLock.Unlock(); log.Printf("requestVoteFromOneServer released lock\n")}()
		}
		if reqTerm != myRaft.state.CurrentTerm{
			return
		}
		switch ret.Success {
		case pb.RaftReturnCode_SUCCESS:
			log.Printf("IN AE -> Append Entry to %s Succeeded : %v\n", serverAddr, e)
			myRaft.nextIndex[serverAddr] += len(sendEntries)
			myRaft.matchIndex[serverAddr] = myRaft.nextIndex[serverAddr] - 1
			if int64(myRaft.matchIndex[serverAddr]) > myRaft.commitIndex &&
				myRaft.state.logs.EntryList[myRaft.matchIndex[serverAddr]].term == myRaft.state.CurrentTerm {
				log.Printf("In AE -> Worth considering matchindex\n")
				if countGreater(myRaft.matchIndex, myRaft.matchIndex[serverAddr]) >= myRaft.majorityNum {
					myRaft.state.PersistentStore()
					myRaft.commitIndex = int64(myRaft.matchIndex[serverAddr])
					log.Printf("In AE -> update commitindex matchindex\n")
					for i := myRaft.lastApplied + 1; i <= myRaft.commitIndex; i++ {

						log.Printf("In AE -> And I entered the for loop\n")
						myRaft.out.ParseAndApplyEntry(myRaft.state.logs.EntryList[i])
						myRaft.lastApplied++
						if myRaft.state.logs.EntryList[i].applyChan != nil {
							myRaft.state.logs.EntryList[i].applyChan <- true
							close(myRaft.state.logs.EntryList[i].applyChan)
							myRaft.state.logs.EntryList[i].applyChan = nil
						}
						log.Printf("In AE -> My applymsg to %v was accepted\n", myRaft.state.logs.EntryList[i].applyChan)
					}
				}
			}
		case pb.RaftReturnCode_FAILURE_TERM:
			log.Printf("IN AE -> Before convertToFollower")
			myRaft.changeToFollower(ret.Term)
			dropAndSet(myRaft.leaderToFollowerChan)
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			myRaft.nextIndex[serverAddr]--
		}
	}
	return
}

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteChan chan bool, reqTerm int64) {
	log.Printf("IN RV -> Send RequestVote to Server: %s\n", serverAddr)

	//if uselock{
	//	myRaft.stateLock.RLock()
	//}
	if uselock{
		myRaft.stateLock.RLock()
		log.Printf("requestVoteFromOneServer acquired Rlock\n")
		//defer func(){myRaft.stateLock.Unlock(); log.Printf("requestVoteFromOneServer released lock\n")}()
	}
	lastLogIndex := int64(len(myRaft.state.logs.EntryList) - 1)
	lastLogTerm := int64(-1)
	if lastLogIndex != -1 {
		lastLogTerm = myRaft.state.logs.EntryList[lastLogIndex].term
	}


	senderId, convErr := strconv.Atoi(myRaft.config.ID)
	if convErr != nil{
		log.Printf("cant convert ID\n")
		//if uselock{
		//	myRaft.stateLock.RUnlock()
		//}
		return
	}

	req := &pb.RVRequest{Term: myRaft.state.CurrentTerm, CandidateID: myRaft.config.ID,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm, Sender:int64(senderId)}

	if uselock{
		myRaft.stateLock.RUnlock()
		log.Printf("requestVoteFromOneServer released Rlock\n")
	}


	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	defer connManager.gc()
	//var e error


	ret, e := connManager.rpcCaller.RequestVote(connManager.ctx, req)
	if e != nil {
		log.Printf("IN RV -> RPC RV ERROR: %v\n", e)
		return
	}

	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("requestVoteFromOneServer acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("requestVoteFromOneServer released lock\n")}()
	}
	if req.Term != myRaft.state.CurrentTerm || myRaft.membership != Candidate{
		return
	}

	if ret.Term > myRaft.state.CurrentTerm {
		log.Printf("Before changeToFollower\n")
		myRaft.changeToFollower(ret.Term)
		dropAndSet(myRaft.leaderToFollowerChan)
		log.Printf("IN RV -> Got Higher Term %d from %s, convert to Follower\n", myRaft.state.CurrentTerm, serverAddr)
	} else {
		log.Printf("Before send vote to countVoteChan\n")
		if countVoteChan != nil {
			 countVoteChan <- ret.VoteGranted
		}
		log.Printf("IN RV -> Got Vote %t from %s\n", ret.VoteGranted, serverAddr)

	}
	return
}

func (myRaft *RaftService) leaderInitVolatileState() {
	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("leaderInitVolatileState acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("leaderInitVolatileState released lock\n")}()
	}
	myRaft.nextIndex = make(map[string]int)
	myRaft.matchIndex = make(map[string]int)
	for _, server := range myRaft.config.ServerList.Servers {
		myRaft.nextIndex[server.Addr] = len(myRaft.state.logs.EntryList)
		myRaft.matchIndex[server.Addr] = -1
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
			retVal := myRaft.appendHeartbeatEntryToOneFollower(server.Addr, myRaft.state.CurrentTerm)
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
			log.Printf("ConfirmLeadership Timeout!")
			confirmationChan <- false
			break Done
		}
	}
}


func (myRaft *RaftService) changeToFollower(term int64) {
	myRaft.membership = Follower
	myRaft.state.CurrentTerm = term
	myRaft.state.VoteFor = ""
	myRaft.state.PersistentStore()
}

func dropAndSet(ch chan bool){
	select {
	case <-ch:
	default:
	}
	ch <- true
}