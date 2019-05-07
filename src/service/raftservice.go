package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	pb "proto"
	"strconv"
	"sync"
	"sync/atomic"
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
	rpcMethodLock     *sync.RWMutex
	nextIndexLock     *sync.Mutex
	matchIndexLock    *sync.Mutex
	stateLock		  *sync.RWMutex
	monkey 		      *MonkeyService
	electionTimer     *time.Timer
	winElectionChan	  chan bool
}

func NewRaftService(appendChan chan entry, out OutService, ID string) *RaftService {
	state := InitState() //todo
	membership := Follower
	heartbeatChan := make(chan bool)
	convertToFollower := make(chan bool)
	config := createConfig(ID) //todo
	majorityNum := config.ServerList.ServerNum/2 + 1
	commitIndex := int64(-1)
	lastApplied := int64(-1)
	rpcMethodLock := new(sync.RWMutex)
	nextIndexLock := &sync.Mutex{}
	matchIndexLock := &sync.Mutex{}
	stateLock := new(sync.RWMutex)
	selfID, _ := strconv.ParseInt(config.ID, 10, 32)
	monkey := NewMonkeyService(config.ServerList.ServerNum, int32(selfID))
	winElectionChan := make(chan bool, 1)
	return &RaftService{state: state, membership: membership, heartbeatChan: heartbeatChan,
		convertToFollower: convertToFollower, config: config, majorityNum: majorityNum, commitIndex: commitIndex,
		lastApplied: lastApplied, appendChan: appendChan, out: out, rpcMethodLock: rpcMethodLock,
		nextIndexLock:nextIndexLock, matchIndexLock:matchIndexLock, stateLock:stateLock, monkey: monkey, winElectionChan:winElectionChan}
}
func (myRaft *RaftService) dropMessage (senderID int64) bool{
	selfID, _:= strconv.ParseInt(myRaft.config.ID, 10, 32)
	log.Printf("senderID is %d, self ID is %d\n", senderID, selfID)
	randNum := rand.Float32()
	probInMat := myRaft.monkey.matrix[senderID][selfID]
	log.Printf("Num in Mat is %f\n", probInMat)
	log.Printf("Generated RandNum is %f\n", randNum)
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
	//select{
	//case myRaft.heartbeatChan <- true:
	//default:
	//}
	myRaft.electionTimer = time.NewTimer(myRaft.randomTimeInterval())
	if myRaft.state.CurrentTerm < req.Term {
		myRaft.state.CurrentTerm = req.Term
		log.Printf("IN RPC AE -> Before send true to convertToFollower")
		myRaft.changeToFollower()
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
		myRaft.state.CurrentTerm = req.Term
		log.Printf("IN RPC RV -> Before send true to convertToFollower")
		myRaft.changeToFollower()
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
	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("leaderAppendEntries acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("leaderAppendEntries released lock\n")}()
	}
	for _, server := range myRaft.config.ServerList.Servers {
		log.Printf("nextIndex of Server %s is %d, myLogLen-1 is %d", server.Addr, myRaft.nextIndex[server.Addr], len(myRaft.state.logs.EntryList)-1)
		if !isFirstHeartbeat && len(myRaft.state.logs.EntryList)-1 >= myRaft.nextIndex[server.Addr] {
			go myRaft.appendEntryToOneFollower(server.Addr)
		}else{
			go myRaft.appendHeartbeatEntryToOneFollower(server.Addr)
		}
	}
}

func (myRaft *RaftService) candidateRequestVotes(quit chan bool) {
	//countVoteCnt := make(chan bool)
	countVoteCnt := int32(1)
	//voteCnt := 1
	for _, server := range myRaft.config.ServerList.Servers {
		go myRaft.requestVoteFromOneServer(server.Addr, &countVoteCnt)
	}
	//voteNum := 0
	//for {
	//	select {
	//	//case vote := <-countVoteChan:
	//	//	voteNum++
	//	//	log.Printf("Collected %d Vote\n", voteNum)
	//	//	if vote {
	//	//		voteCnt++
	//	//	}
	//	//	if voteCnt >= myRaft.majorityNum {
	//	//		winElectionChan <- true
	//	//		//close(countVoteChan)
	//	//		log.Printf("Won Election!!!\n")
	//	//		return
	//	//	}
	//	case <-quit:
	//		//for _, quitForARoutine := range quitForRVRoutines{
	//		//	quitForARoutine <- true
	//		//}
	//		return
	//	//default:
	//	//	log.Printf("Collected %d Vote\n", countVoteCnt)
	//	//	if countVoteCnt >= int32(myRaft.majorityNum){
	//	//		winElectionChan <- true
	//	//		log.Printf("Won Election!!!\n")
	//	//		return
	//	//	}
	//	}
	//}
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
		myRaft.electionTimer = time.NewTimer(myRaft.randomTimeInterval())
		switch myRaft.membership {
		case Leader:
			myRaft.leaderInitVolatileState()
			quit := make(chan bool)
			go myRaft.appendEntriesRoutine(quit)

		Done:
			for {
				select {
				//case <-myRaft.convertToFollower:
				//	myRaft.membership = Follower
				//	if uselock{
				//		myRaft.stateLock.Lock()
				//	}
				//	myRaft.state.VoteFor = ""
				//	myRaft.state.PersistentStore()
				//	if uselock{
				//		myRaft.stateLock.Unlock()
				//	}
				//	quit <- true
				//	break Done
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
				default:
					if myRaft.membership!= Leader{
						quit <- true
						break Done
					}
				}
			}
		case Follower:
			//electionTimer := time.NewTimer(myRaft.randomTimeInterval())
		DoneFollower:
			for{
				select {
				case <-myRaft.electionTimer.C:
					myRaft.membership = Candidate
					break DoneFollower
				//case <-myRaft.heartbeatChan:
				//	break DoneFollower
				//case <-myRaft.convertToFollower:
				//	if uselock{
				//		myRaft.stateLock.Lock()
				//	}
				//	myRaft.state.VoteFor = ""
				//	myRaft.state.PersistentStore()
				//	if uselock{
				//		myRaft.stateLock.Unlock()
				//	}
				}
			}
			//electionTimer.Stop()
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
			//electionTimer := time.NewTimer(myRaft.randomTimeInterval())
			//winElectionChan := make(chan bool)
			quit := make(chan bool)
			go myRaft.candidateRequestVotes(quit)
			select {
			case <-myRaft.electionTimer.C:
				myRaft.membership = Candidate
				quit <- true
			case <-myRaft.winElectionChan:
				myRaft.membership = Leader
			//case <-myRaft.convertToFollower:
			//	myRaft.membership = Follower
			//	quit <- true
			//	if uselock{
			//		myRaft.stateLock.Lock()
			//	}
			//	myRaft.state.VoteFor = ""
			//	myRaft.state.PersistentStore()
			//	if uselock{
			//		myRaft.stateLock.Unlock()
			//	}
			}
			//electionTimer.Stop()

		}
		log.Printf("Membership now: %s, term: %d, voteFor: %s\n\n", nameMap[myRaft.membership],
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

func (myRaft *RaftService) appendHeartbeatEntryToOneFollower(serverAddr string) bool {
	log.Printf("IN HB -> Send Heartbeat to server: %s\n", serverAddr)

	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("appendHeartbeatEntryToOneFollower acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("appendHeartbeatEntryToOneFollower released lock\n")}()
	}
	// Set up a connection to the server.
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
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
	//if uselock{
	//	myRaft.stateLock.RUnlock()
	//}
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	//if uselock{
	//	myRaft.stateLock.Lock()
	//	defer myRaft.stateLock.Unlock()
	//}
	if e != nil {
		log.Printf("IN HB -> Send HeartbeatEntry to %s failed : %v\n", serverAddr, e)
	} else {
		switch ret.Success {
		case pb.RaftReturnCode_SUCCESS:
			log.Printf("IN HB -> heartbeat to %s PREVLOG Success \n", serverAddr)
			myRaft.matchIndex[serverAddr] = myRaft.nextIndex[serverAddr] - 1
			return true
		case pb.RaftReturnCode_FAILURE_TERM:
			log.Printf("IN HB -> heartbeat to %s TERM FAILURE \n", serverAddr)
			myRaft.state.CurrentTerm = ret.Term
			myRaft.state.PersistentStore()
			log.Printf("IN HB -> Before send true to convertToFollower")
			//select{
			//case myRaft.convertToFollower <- true:
			//default:
			//}
			myRaft.changeToFollower()
			return false
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			log.Printf("IN HB -> heartbeat to %s PREVLOG FAILURE \n", serverAddr)
			myRaft.nextIndex[serverAddr]--
			return false
		}
	}
	return false
}

func (myRaft *RaftService) appendEntryToOneFollower(serverAddr string) {
	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("appendEntryToOneFollower acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("appendEntryToOneFollower released lock\n")}()
	}

	//myRaft.stateLock.RLock()
	log.Printf("nextIndex::%v\n", myRaft.nextIndex)
	log.Printf("matchIndex::%v\n", myRaft.matchIndex)
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
	// Set up a connection to the server.
	//myRaft.stateLock.RUnlock()
	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	ret, e := connManager.rpcCaller.AppendEntries(connManager.ctx, req)
	defer connManager.gc()
	//myRaft.stateLock.Lock()
	//defer myRaft.stateLock.Unlock()
	if e != nil {
		log.Printf("IN AE -> Entry to %s failed RPC error : %v\n", serverAddr, e)
	} else {
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
			myRaft.state.CurrentTerm = ret.Term
			myRaft.state.PersistentStore()
			log.Printf("IN AE -> Before send true to convertToFollower")
			//select{
			//case myRaft.convertToFollower <- true:
			//default:
			//}
			myRaft.changeToFollower()
		case pb.RaftReturnCode_FAILURE_PREVLOG:
			myRaft.nextIndex[serverAddr]--
		}
	}
	return
}

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteCnt *int32) {
	log.Printf("IN RV -> Send RequestVote to Server: %s\n", serverAddr)

	if uselock{
		myRaft.stateLock.Lock()
		log.Printf("requestVoteFromOneServer acquired lock\n")
		defer func(){myRaft.stateLock.Unlock(); log.Printf("requestVoteFromOneServer released lock\n")}()
	}

	connManager := createConnManager(serverAddr, time.Duration(myRaft.config.RpcTimeout))
	//if uselock{
	//	myRaft.stateLock.RLock()
	//}
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

	//if uselock{
	//	myRaft.stateLock.RUnlock()
	//}

	defer connManager.gc()
	//var e error


	ret, e := connManager.rpcCaller.RequestVote(connManager.ctx, req)
	if e != nil {
		log.Printf("IN RV -> RPC RV ERROR: %v\n", e)
		return
	}
	//
	//if uselock{
	//	myRaft.stateLock.Lock()
	//	defer myRaft.stateLock.Unlock()
	//}
	if ret.Term > myRaft.state.CurrentTerm {
		myRaft.state.CurrentTerm = ret.Term
		myRaft.state.PersistentStore()
		log.Printf("Before send true to convertToFollower\n")
		//if myRaft.convertToFollower != nil{
		//	select{
		//	case myRaft.convertToFollower <- true:
		//	default:
		//	}
		//}
		myRaft.changeToFollower()
		log.Printf("IN RV -> Got Higher Term %d from %s, convert to Follower\n", myRaft.state.CurrentTerm, serverAddr)
	} else {
		log.Printf("Before send vote to countVoteChan\n")
		//if countVoteChan != nil {
		//	select{
		//	case countVoteChan <- ret.VoteGranted:
		//	default:
		//	}
		//}
		atomic.AddInt32(countVoteCnt, 1)
		*countVoteCnt++
		if *countVoteCnt>=int32(myRaft.majorityNum){
			dropAndSet(myRaft.winElectionChan)
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
			log.Printf("ConfirmLeadership Timeout!")
			confirmationChan <- false
			break Done
		}
	}
}

func (myRaft *RaftService) changeToFollower(){
	log.Printf("Change membership from %v term%v to follower)", nameMap[myRaft.membership],
		myRaft.state.CurrentTerm)
	myRaft.membership = Follower
	myRaft.state.VoteFor = ""
}


func dropAndSet(ch chan bool) {
	select {
	case <-ch:
	default:
	}
	ch <- true
}