package service

import(
	"fmt"
	pb "proto"
)

type Log struct{
	EntryList []entry
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
	var appendIndex int64
	for i, reqEntry := range reqEntries{
		appendIndex = appendStartIndex + int64(i)
		fmt.Printf("look here bitch : %d, len of log: %d", appendStartIndex+int64(i), len(log.EntryList))
		log.EntryList = append(log.EntryList[0:appendIndex], entry{op:reqEntry.Op, val:reqEntry.Val, term:reqEntry.Term})
	}
}

func NewLog() *Log{
	ret := &Log{}
	ret.EntryList = make([]entry, 0)
	ret.commitIndex = -1
	ret.lastApplied = -1
	return ret
}

