package service

import (
	pb_monkey "chaosmonkey"
	"context"
	"fmt"
)

type MonkeyService struct {
	matrix [][]float32
	selfID int32
}


func NewMonkeyService(n int, id int32) *MonkeyService {
	var matrix [][]float32
	for i:= 0; i < n; i++ {
		tmp := make([]float32, n)
		matrix = append(matrix, tmp)
	}
	monkey := &MonkeyService{matrix: matrix, selfID: id}
	return monkey
}



func (s *MonkeyService) KillServer(ctx context.Context, req *pb_monkey.ServerStat) (*pb_monkey.Status, error) {
	serverID := req.ServerID
	for i := 0; i < len(s.matrix); i++ {
		s.matrix[i][serverID] = 1
	}

	for i := 0; i < len(s.matrix[0]); i++ {
		s.matrix[serverID][i] = 1
	}
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
}


func (s *MonkeyService) Partition(ctx context.Context, req *pb_monkey.PartitionInfo) (*pb_monkey.Status, error) {
	partitionServers := req.Server
	local_map := make(map[int]int)
	for _, server := range partitionServers {
		local_map[int(server.ServerID)] = 1
	}
	_, contained := local_map[int(s.selfID)]
	// this server is not in the partition list
	if contained == false {
		for _, server := range partitionServers {
			s.matrix[s.selfID][server.ServerID] = 1 // cannot send messages to server in the partition
			s.matrix[server.ServerID][s.selfID] = 1 // cannot receive messages to server in the partition
		}
	} else { // this server is in the partition list

		for _, server := range partitionServers {
			s.matrix[s.selfID][server.ServerID] = 0
			s.matrix[server.ServerID][s.selfID] = 0
		}

		for i := 0; i < len(s.matrix); i++ {
			_, ok := local_map[i]
			if !ok {
				s.matrix[s.selfID][i] = 1
				s.matrix[i][s.selfID] = 1
			}
		}
	}
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
}
func (s *MonkeyService) UploadMatrix(ctx context.Context, req *pb_monkey.ConnMatrix) (*pb_monkey.Status, error) {
	rows := req.GetRows()
	for i, v := range rows {
		s.matrix[i] = v.GetVals()
	}
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
}

func (s *MonkeyService) UpdateValue(ctx context.Context, req *pb_monkey.MatValue) (*pb_monkey.Status, error) {
	row := req.Row
	col := req.Col
	value := req.Val
	s.matrix[row][col] = value
	ret := &pb_monkey.Status{Ret: pb_monkey.StatusCode_OK}
	for _, v := range s.matrix {
		for _, k:= range v {
			fmt.Print(k)
			fmt.Print(" ")
		}
		fmt.Println(" ")
	}
	fmt.Println(" ")
	return ret, nil
}