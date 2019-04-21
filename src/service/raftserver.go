package service


type RaftService struct{

}

func (*RaftService) AppendEntries(){

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

