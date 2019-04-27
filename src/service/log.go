package service

import(
	pb "proto"
)

type Log struct{
	EntryList []entry
	lastIndex int64 // initial value is 1 not 0 according to paper
	commitIndex int64
	lastApplied int64

}

type entry struct{
	op string //put only?
	key string
	val string // e.g. op=put val=key1:v1
	term int64
}

func (log *Log) cutEntries(cutStartIndex int64){
	log.EntryList = log.EntryList[0:cutStartIndex]
}

func (log *Log) appendEntries(appendStartIndex int64, reqEntries []*pb.Entry){
	for i, reqEntry := range reqEntries{
		log.EntryList[appendStartIndex+int64(i)] = entry{op:reqEntry.Op, val:reqEntry.Val, term:reqEntry.Term}
	}
}

func (log *Log) getUnappliedEntries() []entry{
	//var unAppliedEntries []entry
	//if log.lastApplied < log.commitIndex{
	//	unAppliedEntries = log.EntryList[log.lastApplied+1 : log.commitIndex+1]
	//	return unAppliedEntries
	//}else{
	//	return nil
	//}
}
