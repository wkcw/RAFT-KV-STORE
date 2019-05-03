package service

import (
	pb "proto"
)

func entryToPbentry(myEntry entry) *pb.Entry{
	ret := new(pb.Entry)

	ret.Op = myEntry.op
	ret.Key = myEntry.key
	ret.Val = myEntry.val
	ret.Term = myEntry.term

	return ret
}

func maxIntIndex(a []int) int{
	maxVal := -1000000000
	maxIndex := -1
	for i, val := range a{
		if val > maxVal{
			maxIndex = i
			maxVal = val
		}
	}
	return maxIndex
}

func countGreater(a map[string]int, num int) int{
	cnt := 1
	for _, val := range a{
		if val >= num{
			cnt++
		}
	}
	return cnt
}

func swap(a []int, i int, j int){
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}

type SerialPair struct{
	SerialNo int64
	Response pb.PutResponse
}