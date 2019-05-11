package test

import (
	pb_monkey "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	pb "proto"
	"testing"
	"time"
	"util"
)

type PartitionParams struct {
	size int
	servers []*pb_monkey.Server
}

func generateRandomList(serverNum int) []int{
	arr := make([]int, serverNum)
	for i := 0; i < serverNum; i++ {
		arr[i] = i;
	}
	rand.Seed(time.Now().Unix())
	for i := len(arr) - 1; i >= 0; i-- {
		num := rand.Intn(len(arr))
		arr[i], arr[num] = arr[num], arr[i]
	}

	return arr
}

func generatePartitionParams(serverNum int, randomIdList []int, partitionSize int) *PartitionParams {
	servers := make([]*pb_monkey.Server, 0)
	fmt.Println("Partition group is: ")
	for i := 0; i < partitionSize; i++ {
		var server pb_monkey.Server
		server.ServerID = int32(randomIdList[i])
		fmt.Printf("%d ", randomIdList[i])
		servers = append(servers, &server)
	}
	return &PartitionParams{size: partitionSize, servers: servers}
}

func checkLeader(params *PartitionParams) int {
	config := util.CreateConfig()
	serverlist := config.ServerList
	var count int
	// check partitioned group
	for _, server := range params.servers {
		address := serverlist.Servers[server.ServerID].Addr
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewKeyValueStoreClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// generating random params for
		response, _ := client.IsLeader(ctx, &pb.CLRequest{})
		if response.IsLeader == true {
			count++
		}
	}

	fmt.Printf("Leader number in the partition is: %d\n", count)
	return count
}

func clear() {
	config := util.CreateConfig()
	serverlist := config.ServerList
	for _, server := range serverlist.Servers {
		var address string = server.Addr
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb_monkey.NewChaosMonkeyClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		matrix := util.CreateConnMatrix(serverlist.ServerNum)
		matrows := make([]pb_monkey.ConnMatrix_MatRow, serverlist.ServerNum)
		for i := 0; i < len(matrows); i++ {

			matrows[i].Vals = matrix[i]
		}

		matrows_ptr := make([]*pb_monkey.ConnMatrix_MatRow, serverlist.ServerNum)
		for i := 0; i < len(matrows_ptr); i++ {
			matrows_ptr[i] = &matrows[i]
		}
		c.UploadMatrix(ctx, &pb_monkey.ConnMatrix{Rows: matrows_ptr})
	}
}
// Randomized one majority parition, leader must exist in the partition
func Test_Parition_1(t *testing.T) {
	clear()
	time.Sleep(500 * time.Millisecond)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2 + 1)
	for _, server := range config.ServerList.Servers {
		var address string = server.Addr
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb_monkey.NewChaosMonkeyClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// generating random params for
		client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	}

	time.Sleep(500 * time.Millisecond)


	leaderCount := checkLeader(params)

	if leaderCount == 1 {
		t.Log("Passed the majority test.")
	} else {
		t.Error("Failed the majority test.")
	}
	//clear()
}

// Randomized one minority parition, leader must not exist in the partition
func Test_Parition_2(t *testing.T) {
	clear()
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2)
	for _, server := range config.ServerList.Servers {
		var address string = server.Addr
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb_monkey.NewChaosMonkeyClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// generating random params for
		client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	}

	time.Sleep(750)


	leaderCount := checkLeader(params)

	if leaderCount <= 1 {
		t.Log("Passed the minority test.")
	} else {
		t.Error("Failed the minority test.")
	}
	clear()
}
