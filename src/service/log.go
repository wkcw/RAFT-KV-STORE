package service

type Log struct{
	EntryList []entry
	lastIndex int // initial value is 1 not 0 according to paper
}

type entry struct{
	op string //put only?
	val string // e.g. op=put val=key1:v1
	term int
}

func (log *Log) cutEntries(cutStartIndex int){
	log.EntryList = log.EntryList[0:cutStartIndex]
}