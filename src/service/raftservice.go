package service

import (
	pb "proto"
	"context"
	"time"
	"log"
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
}

func NewRaftService() *RaftService{
	return &RaftService{}
}


func (myRaft *RaftService) AppendEntries(ctx context.Context, req pb.AERequest) (*pb.AEResponse, error){

	response := &pb.AEResponse{Term:myRaft.state.CurrentTerm}

	//rule 1
	if req.Term < myRaft.state.CurrentTerm {
		response.Success = false
		return response, nil
	}

	myRaft.heartbeatChan <- true
	if myRaft.state.CurrentTerm < req.Term{
		myRaft.state.CurrentTerm = req.Term
		myRaft.convertToFollower <- true
	}
	//rule 2
	if len(myRaft.state.logs.EntryList) < int(req.PrevLogIndex) || myRaft.state.logs.EntryList[req.PrevLogIndex-1].term != req.PrevLogTerm{
		response.Success = false
		return response, nil

	}
	appendStartIndex := int64(req.PrevLogIndex)
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
		myRaft.state.logs.commitIndex = minInt64(req.LeaderCommit, int64(len(myRaft.state.logs.EntryList)))
	}
	response.Success = true
	return response, nil

}

func (myRaft *RaftService) RequestVote(ctx context.Context, req *pb.RVRequest) (*pb.RVResponse, error){
	if myRaft.state.CurrentTerm < req.Term{
		myRaft.state.CurrentTerm = req.Term
	}
	// reply false if term < currentTerm
	if req.Term < myRaft.state.CurrentTerm {
		return &pb.RVResponse{Term: myRaft.state.CurrentTerm, VoteGranted: false}, nil
	}
	// candidate's log is at least as up-to-date as receiver's log?
	if myRaft.state.VoteFor == "" || myRaft.state.VoteFor == req.CandidateID {
		return &pb.RVResponse{Term: req.Term, VoteGranted: true}, nil
	}
	return &pb.RVResponse{Term: req.Term, VoteGranted: false}, nil
}

func (myRaft *RaftService) leaderAppendHeartbeatEntries(){
	for _, serverAddr := range myRaft.config.peers{
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
	for {
		select {
		case <- timeoutTicker.C:
			go myRaft.leaderAppendHeartbeatEntries()
		case <- quit:
			timeoutTicker.Stop()
			return
		}
	}
}

func (myRaft * RaftService) randomTimeInterval() time.Duration{
	return 50 * time.Millisecond
}

func (myRaft *RaftService) serverMain(){
	for {

		switch myRaft.membership {
			case Leader:
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
				}
			case Candidate:
				myRaft.state.CurrentTerm++;
				myRaft.state.VoteFor = myRaft.config.ID
				electionTimer := time.NewTimer(myRaft.randomTimeInterval())
				winElectionChan := make(chan bool)
				go myRaft.candidateRequestVotes(winElectionChan)
				select {
				case <- electionTimer.C:
					myRaft.membership = Candidate

				case <- winElectionChan:
					electionTimer.Stop()
				}

		}
	}
}


//get server state
//func (*RaftService) checkAll(){
//	return all states
//}

//func (*RaftService) checkLog(){
//	return log
//}
//
//func (*RaftService) checkIdentity(){
//	return its identity
//}
//
//func (*RaftService) checkVoteFor(){
//	return who it has voted for
//}
//
//func (*RaftService) checkGotVotes(){
//	return who has voted for it and who are they
//}
//
//func (*RaftService) checkTerm(){
//	return term
//}


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

func (myRaft *RaftService) requestVoteFromOneServer(serverAddr string, countVoteChan chan bool){
	connManager := createConnManager(serverAddr)
	req := &pb.RVRequest{Term:myRaft.state.CurrentTerm, CandidateID:myRaft.config.ID,
							LastLogIndex:myRaft.state.logs.lastIndex,
							LastLogTerm:myRaft.state.logs.EntryList[myRaft.state.logs.lastIndex].term}
	ret, e := connManager.rpcCaller.RequestVote(connManager.ctx, req)
	defer connManager.gc()
	if e!=nil{
		log.Printf("Send RequestVote to %s failed : %v\n", serverAddr, e)
	}else{
		if ret.Term>myRaft.state.CurrentTerm{
			myRaft.state.CurrentTerm = ret.Term
			myRaft.convertToFollower <- true
		}else{
			countVoteChan <- ret.VoteGranted
		}
	}
	return
}
