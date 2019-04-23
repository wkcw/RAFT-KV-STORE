package service

import(
	pb "proto"
	"context"
)

type RaftService struct{
	currentTerm int64
	log Log
}


func (myRaft *RaftService) AppendEntries(ctx context.Context, req pb.AERequest) (*pb.AEResponse, error){
	response := &pb.AEResponse{Term:myRaft.currentTerm}
	//rule 1
	if req.Term < myRaft.currentTerm {
		response.Success = false
		return response, nil
	}
	//rule 2
	if len(myRaft.log.EntryList) < int(req.PrevLogIndex) || myRaft.log.EntryList[req.PrevLogIndex-1].term != req.PrevLogTerm{
		response.Success = false
		return response, nil

	}
	appendStartIndex := int64(req.PrevLogIndex)
	//rule 3
	for _, reqEntry := range req.Entries{
		if myRaft.log.EntryList[appendStartIndex].term != reqEntry.Term {
			myRaft.log.cutEntries(appendStartIndex)
			break
		}
		appendStartIndex++
	}
	//rule 4
	myRaft.log.appendEntries(appendStartIndex, req.Entries)
	//rule 5
	if req.LeaderCommit > myRaft.log.commitIndex {
		myRaft.log.commitIndex = minInt64(req.LeaderCommit, int64(len(myRaft.log.EntryList)))
	}
	response.Success = true
	return response, nil

}

func (*RaftService) RequestVote(){

}

func (*RaftService)




//get server state
func (*RaftService) checkAll(){
	return all states
}

func (*RaftService) checkLog(){
	return log
}

func (*RaftService) checkIdentity(){
	return its identity
}

func (*RaftService) checkVoteFor(){
	return who it has voted for
}

func (*RaftService) checkGotVotes(){
	return who has voted for it and who are they
}

func (*RaftService) checkTerm(){
	return term
}


func minInt64(num1 int64, num2 int64) int64{
	if num1<num2 {
		return num1
	}else{
		return num2
	}
}

