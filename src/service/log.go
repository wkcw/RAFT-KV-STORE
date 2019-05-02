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
	applyChan chan bool
}

func (log *Log) cutEntries(cutStartIndex int64){
	for i:=cutStartIndex; i<int64(len(log.EntryList)); i++{
		log.EntryList[i].applyChan <- false
		close(log.EntryList[i].applyChan)
		log.EntryList[i].applyChan = nil
	}
	log.EntryList = log.EntryList[0:cutStartIndex]
}

func (log *Log) appendEntries(appendStartIndex int64, reqEntries []*pb.Entry){
	for i, reqEntry := range reqEntries{
		log.EntryList[appendStartIndex+int64(i)] = entry{op:reqEntry.Op, val:reqEntry.Val, term:reqEntry.Term}
	}
}

