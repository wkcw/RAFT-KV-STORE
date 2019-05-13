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
	containLeader bool
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

func generatePartitionParams(serverNum int, randomIdList []int, partitionSize int, leaderID int32) *PartitionParams {
	servers := make([]*pb_monkey.Server, 0)
	var containLeader bool = false
	fmt.Println("Partition group is: ")
	for i := 0; i < partitionSize; i++ {
		var server pb_monkey.Server
		server.ServerID = int32(randomIdList[i])
		if server.ServerID == leaderID {
			containLeader = true
		}
		fmt.Printf("%d ", randomIdList[i])
		servers = append(servers, &server)
	}
	return &PartitionParams{size: partitionSize, servers: servers, containLeader:containLeader}
}
func generateMultiPartitionParams(serverNum int, randomIdList []int, partitionSize int, leaderID int32) []*PartitionParams {
	partitionParamsList := make([]*PartitionParams, 0)
	totalNum := serverNum / partitionSize
	fmt.Printf("total partition group number is: %d\n", totalNum)
	index := 0
	for i := 0; i < totalNum; i++ {
		fmt.Printf("Now partitioning Group %d\n", i)
		servers := make([]*pb_monkey.Server, 0)
		var containLeader bool = false
		fmt.Printf("Parition size is: %d\n", partitionSize)
		for j := 0; j < partitionSize; j++ {
			var server pb_monkey.Server
			server.ServerID = int32(randomIdList[index])
			fmt.Printf("Server ID is %d, index is %d\n", server.ServerID, index)
			index++
			if server.ServerID == leaderID {
				containLeader = true
			}
			servers = append(servers, &server)
		}
		params := &PartitionParams{size: partitionSize, servers: servers, containLeader:containLeader}
		partitionParamsList = append(partitionParamsList, params)
	} // 2 * 2

	// last partition has 1 server
	fmt.Printf("Now partitioning Group %d\n", totalNum)
	servers := make([]*pb_monkey.Server, 0)
	var containLeader bool = false
	var server pb_monkey.Server
	server.ServerID = int32(randomIdList[index])
	fmt.Printf("Server ID is %d, index is %d\n", server.ServerID, index)
	if server.ServerID == leaderID {
		containLeader = true
	}
	servers = append(servers, &server)
	params := &PartitionParams{size: partitionSize, servers: servers, containLeader:containLeader}
	partitionParamsList = append(partitionParamsList, params)

	return partitionParamsList
}
func partitionLeader(leaderID int32) *PartitionParams {
	servers := make([]*pb_monkey.Server, 0)
	server := &pb_monkey.Server{ServerID:leaderID}
	servers = append(servers, server)
	params := &PartitionParams{size: 1, servers:servers}
	return params
}
func checkLeaderFromAllServers(serverList util.ServerList) int{
	var count int = 0
	for _, server := range serverList.Servers {
		conn, err := grpc.Dial(server.Addr, grpc.WithInsecure())
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
	return count
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

	return count
}

func clear() {
	config := util.CreateConfig()
	serverlist := config.ServerList
	doneChan := make(chan bool, serverlist.ServerNum)
	for _, server := range serverlist.Servers {
		go clearOne(server.Addr, serverlist.ServerNum, doneChan)
	}
	for i:=0; i<serverlist.ServerNum; i++{
		<- doneChan
	}
	return
}

func clearOne(address string, size int, doneChan chan bool){
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb_monkey.NewChaosMonkeyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	matrows := make([]pb_monkey.ConnMatrix_MatRow, size)
	for i := 0; i < len(matrows); i++ {
		matrows[i].Vals = make([]float32, size)
	}

	matrows_ptr := make([]*pb_monkey.ConnMatrix_MatRow, size)
	for i := 0; i < len(matrows_ptr); i++ {
		matrows_ptr[i] = &matrows[i]
	}
	c.UploadMatrix(ctx, &pb_monkey.ConnMatrix{Rows: matrows_ptr})
	doneChan <- true
}

func findLeaderID() int32 {
	config := util.CreateConfig()
	serverlist := config.ServerList
	leaderIDChan := make(chan int32, serverlist.ServerNum)
	for _, server := range serverlist.Servers {
		go askLeaderIDToOne(server, leaderIDChan)
	}
	countMajorArray := make([]int32, serverlist.ServerNum)
	for i:=0; i < serverlist.ServerNum; i++{
		 countMajorArray[<- leaderIDChan]++
	}
	for i:=0; i < serverlist.ServerNum; i++{
		if countMajorArray[i] >= int32(serverlist.ServerNum / 2 + 1){
			return int32(i)
		}
	}
	return -1
}

func askLeaderIDToOne(server util.Server, leaderIDChan chan int32){
	// connect to the server
	conn, err := grpc.Dial(server.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewKeyValueStoreClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	response, _ := c.IsLeader(ctx, &pb.CLRequest{})

	if response.IsLeader == true {
		leaderID := int32(server.ServerId)
		fmt.Println("Catched a leader!")
		leaderIDChan <- leaderID
		return
	}else{
		leaderIDChan <- response.LeaderId
		return
	}
}

// Randomized one majority partition, leader must exist in the partition
func Test_Partition_1(t *testing.T) {
	clear() // reset to 0
	time.Sleep(time.Millisecond * 2000)
	leaderID := findLeaderID() //useless in this test case
	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)

	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2 + 1, leaderID)
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
	time.Sleep(time.Millisecond * 2000)
	leaderCount := checkLeader(params)
	fmt.Printf("Leader number in this partition is: %v", leaderCount)
	if leaderCount == 1 {
		t.Log("Passed the majority test.")
	} else {
		t.Error("Failed the majority test.")
	}
}

// Randomized one minority partition, leader must not exist in the partition
func Test_Partition_2(t *testing.T) {
	clear()
	time.Sleep(time.Millisecond * 10000)
	leaderID := findLeaderID()
	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2, leaderID)
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
	time.Sleep(time.Millisecond * 10000)
	leaderCount := checkLeader(params)

	if (leaderCount == 1 && params.containLeader == true) || (leaderCount == 0 && params.containLeader == false){
		t.Log("Passed the minority test.")
	} else {
		t.Error("Failed the minority test.")
	}
}

// testing all groups are minority, no leader elected at all time
func Test_Partition_3(t *testing.T) {
	clear()
	time.Sleep(time.Millisecond * 3000)
	leaderID := findLeaderID()

	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	paramsList := generateMultiPartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2, leaderID)

	for _, params := range paramsList {
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
	}

	time.Sleep(time.Millisecond * 3000)

	var pass bool = true
	for i, params := range paramsList {
		leaderCount := checkLeader(params)
		fmt.Printf("Leader number in the partition %d is: %d\n", i, leaderCount)
		if params.containLeader == false && leaderCount != 0 {
			pass = false
		}
	}

	if pass {
		t.Log("Passed the multi-minority-partition test")
	} else {
		t.Error("Failed the multi-minority-partition test")
	}
}


//// test how long the cluster would take to become stable after the partition
//// partition majority + minority
//func Test_Partition_Performance_1(t *testing.T) {
//	clear()
//	time.Sleep(time.Millisecond * 2000)
//	leaderID := findLeaderID()
//	fmt.Printf("Current leader is: %d\n", leaderID)
//
//
//	config := util.CreateConfig()
//	serverlist := config.ServerList
//	params := partitionLeader(leaderID)
//	// kill the leader
//	for _, server := range serverlist.Servers {
//		conn, err := grpc.Dial(server.Addr, grpc.WithInsecure())
//		if err != nil {
//			log.Fatalf("did not connect: %v", err)
//		}
//		defer conn.Close()
//		client := pb_monkey.NewChaosMonkeyClient(conn)
//		// Contact the server and print out its response.
//		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
//		defer cancel()
//
//		client.Partition(ctx, &pb_monkey.PartitionInfo{Server:params.servers})
//	}
//	//timer.start()
//
//	for {
//		leaderCount := checkLeader()
//
//		if leaderCount == 1 {
//			timer.stop()
//			break;
//		}
//	}
//
//}
//
//
//
//
