package service
import(
	pb "proto"
)

func entryToPbentry(myEntry entry) *pb.AERequest_Entry{
	ret := &pb.AERequest_Entry{myEntry.op, myEntry.key, myEntry.val, myEntry.term}
	return pb.Entry{myEntry.op, myEntry.key, myEntry.val, myEntry.term}
}