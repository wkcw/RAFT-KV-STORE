package service
import(
	pb "proto"
)

func entryToPbentry(myEntry entry) *pb.AERequest_Entry{
	ret := &pb.AERequest_Entry{myEntry.op, myEntry.key, myEntry.val, myEntry.term}
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
	cnt := 0
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