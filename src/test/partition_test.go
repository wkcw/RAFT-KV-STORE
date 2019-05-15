package test

import (
	pb_monkey "chaosmonkey"
	"client"
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
func pickRandomServerExcludingLeader(excludedServerID int32, serverNum int) int32 {
	rand.Seed(time.Now().UnixNano())
	randomID := rand.Int31n(int32(serverNum - 1))
	if randomID >= excludedServerID {
		randomID++
	}
	fmt.Printf("Choosing server ID: %d\n", randomID)
	return randomID
}

func pickRandomServer(serverNum int) int32{
	rand.Seed(time.Now().UnixNano())
	return rand.Int31n(int32(serverNum))
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
func partitionLeader(leaderID int32) []*PartitionParams {
	config := util.CreateConfig()
	serverlist := config.ServerList
	servers := make([]*pb_monkey.Server, 0)
	server := &pb_monkey.Server{ServerID:leaderID}
	servers = append(servers, server)
	params := &PartitionParams{size: 1, servers:servers}

	serversAnother := make([]*pb_monkey.Server, 0)
	for _, server := range serverlist.Servers {
		if int32(server.ServerId) == leaderID {
			continue
		}
		server := &pb_monkey.Server{ServerID: int32(server.ServerId)}
		serversAnother = append(serversAnother, server)
	}
	paramsAnother := &PartitionParams{size: serverlist.ServerNum - 1, servers: serversAnother}

	result := make([]*PartitionParams, 0)
	result = append(result, params)
	result = append(result, paramsAnother)
	return result
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

func parallel_clear() {
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


func clear() {
	config := util.CreateConfig()
	serverlist := config.ServerList
	for _, server := range serverlist.Servers {
		conn, err := grpc.Dial(server.Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb_monkey.NewChaosMonkeyClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		matrows := make([]pb_monkey.ConnMatrix_MatRow, serverlist.ServerNum)
		for i := 0; i < len(matrows); i++ {
			matrows[i].Vals = make([]float32, serverlist.ServerNum)
		}

		matrows_ptr := make([]*pb_monkey.ConnMatrix_MatRow, serverlist.ServerNum)
		for i := 0; i < len(matrows_ptr); i++ {
			matrows_ptr[i] = &matrows[i]
		}
		c.UploadMatrix(ctx, &pb_monkey.ConnMatrix{Rows: matrows_ptr})
		conn.Close()
		cancel()
	}
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
	response, isLeaderErr := c.IsLeader(ctx, &pb.CLRequest{})
	if isLeaderErr != nil{
		log.Printf("RPC ERR NOT NIL -> %v\n", isLeaderErr)
	}

	if response.IsLeader == true {
		leaderID := int32(server.ServerId)
		fmt.Println("Catched a leader!")
		fmt.Printf("\n")
		leaderIDChan <- leaderID
		return
	}else{
		leaderIDChan <- response.LeaderId
		return
	}
}

// Randomized one majority partition, leader must exist in the partition
func Test_Partition_1(t *testing.T) {
	//clear() // reset to 0
	parallel_clear()
	time.Sleep(time.Millisecond * 2000)
	leaderID := findLeaderID() //useless in this test case
	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)

	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2 + 1, leaderID)
	uploadToOnePartition(params, serverlist)
	//for _, server := range config.ServerList.Servers {
	//	var address string = server.Addr
	//	// connect to the server
	//	conn, err := grpc.Dial(address, grpc.WithInsecure())
	//	if err != nil {
	//		log.Fatalf("did not connect: %v", err)
	//	}
	//	client := pb_monkey.NewChaosMonkeyClient(conn)
	//	// Contact the server and print out its response.
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//	// generating random params for
	//	client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	//	conn.Close()
	//	cancel()
	//}
	time.Sleep(time.Millisecond * 3000)
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
	//clear()
	parallel_clear()
	time.Sleep(time.Millisecond * 10000)
	leaderID := findLeaderID()
	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	params := generatePartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2, leaderID)
	uploadToOnePartition(params, serverlist)
	//for _, server := range config.ServerList.Servers {
	//	var address string = server.Addr
	//	// connect to the server
	//	conn, err := grpc.Dial(address, grpc.WithInsecure())
	//	if err != nil {
	//		log.Fatalf("did not connect: %v", err)
	//	}
	//	client := pb_monkey.NewChaosMonkeyClient(conn)
	//	// Contact the server and print out its response.
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//	// generating random params for
	//	client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	//
	//	conn.Close()
	//	cancel()
	//}
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
	//clear()
	parallel_clear()
	time.Sleep(time.Millisecond * 3000)
	leaderID := findLeaderID()

	fmt.Printf("Current leader is: %d\n", leaderID)
	config := util.CreateConfig()
	serverlist := config.ServerList
	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	paramsList := generateMultiPartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2, leaderID)
	uploadToOnePartitionList(paramsList, serverlist)
	//for _, params := range paramsList {
	//	for _, server := range config.ServerList.Servers {
	//		var address string = server.Addr
	//		// connect to the server
	//		conn, err := grpc.Dial(address, grpc.WithInsecure())
	//		if err != nil {
	//			log.Fatalf("did not connect: %v", err)
	//		}
	//		client := pb_monkey.NewChaosMonkeyClient(conn)
	//		// Contact the server and print out its response.
	//		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//		// generating random params for
	//		client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	//		conn.Close()
	//		cancel()
	//	}
	//}

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


// test how long the cluster would take to become stable after the partition
// partition majority + minority
func Test_Partition_Performance_1(t *testing.T) {
	//clear()
	parallel_clear()
	time.Sleep(time.Millisecond * 5000)
	leaderID := findLeaderID()
	fmt.Printf("Current leader is: %d\n", leaderID)

	// partition leader
	config := util.CreateConfig()
	serverlist := config.ServerList
	paramsList := partitionLeader(leaderID)

	uploadToOnePartitionList(paramsList, serverlist)

	tStart := time.Now()
	var pickedID int32
	pickedID = pickRandomServerExcludingLeader(leaderID, serverlist.ServerNum) // pick one random server from the remaining cluster

	// put to the remaining servers
	for {
		// connect to the server
		conn, err := grpc.Dial(serverlist.Servers[pickedID].Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: "reconnect", Value: "leaderout"})
		if errCode != nil {
			log.Fatalf("could not put raft, an timeout occurred: %v", errCode)
		}

		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			reconnLeaderID := response.LeaderID
			if reconnLeaderID == leaderID || reconnLeaderID == -1 {
				pickedID = pickRandomServerExcludingLeader(leaderID, serverlist.ServerNum)
			} else {
				pickedID = reconnLeaderID
			}
			conn.Close()
			cancel()
			continue;
		}

		if response.Ret == pb.ReturnCode_SUCCESS {
			elapsed := time.Since(tStart)
			log.Println("Put to the leader successfully")
			log.Printf("Elapsed time: %v", elapsed)
			conn.Close()
			cancel()
			break;
		}
		conn.Close()
		cancel()
	}
}


// partition the cluster into all-minored groups
// and then bring the cluster back online
// and put
func Test_Partition_Performance_2(t *testing.T) {
	//clear()
	parallel_clear()
	time.Sleep(time.Millisecond * 3000)
	leaderID := findLeaderID()
	fmt.Printf("Current leader is: %d\n", leaderID)

	config := util.CreateConfig()
	serverlist := config.ServerList

	IdList := generateRandomList(serverlist.ServerNum)
	// majority partition
	paramsList := generateMultiPartitionParams(serverlist.ServerNum, IdList, serverlist.ServerNum / 2, leaderID)

	uploadToOnePartitionList(paramsList, serverlist)
	//for _, params := range paramsList {
	//	for _, server := range config.ServerList.Servers {
	//		var address string = server.Addr
	//		// connect to the server
	//		conn, err := grpc.Dial(address, grpc.WithInsecure())
	//		if err != nil {
	//			log.Fatalf("did not connect: %v", err)
	//		}
	//		defer conn.Close()
	//		client := pb_monkey.NewChaosMonkeyClient(conn)
	//		// Contact the server and print out its response.
	//		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//		defer cancel()
	//		// generating random params for
	//		client.Partition(ctx, &pb_monkey.PartitionInfo{Server: params.servers})
	//	}
	//}

	//wait for more than max election timeout
	time.Sleep(2 * time.Second)


	// remove the partition
	//clear()
	parallel_clear()

	tStart := time.Now()


	var pickedID int32
	pickedID = pickRandomServer(serverlist.ServerNum)
	for {
		// connect to the server
		conn, err := grpc.Dial(serverlist.Servers[pickedID].Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: "performance_test_key", Value: "performance_test_value"})
		if errCode != nil {
			log.Printf("could not put raft, an timeout occurred: %v", errCode)
			pickedID = pickRandomServer(serverlist.ServerNum)
			continue
		}


		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			reconnLeader := response.LeaderID
			if reconnLeader == -1{
				pickedID = pickRandomServer(serverlist.ServerNum)
			} else{
				pickedID = reconnLeader
			}

			conn.Close()
			cancel()
			continue;
		}

		if response.Ret == pb.ReturnCode_SUCCESS {
			elapsed := time.Since(tStart)
			log.Println("Put to the leader successfully")
			log.Printf("Time elapsed: %v", elapsed)
			conn.Close()
			cancel()
			break;
		}

		conn.Close()
		cancel()
	}

}


//Concurrent Upload Start
func uploadToOnePartitionList(paramsList []*PartitionParams, serverlist util.ServerList){
	doneChan := make(chan bool, len(paramsList))
	// partition the cluster into all-minored groups
	for i:=0; i<len(paramsList); i++{
		go func(i int){
			uploadToOnePartition(paramsList[i], serverlist)
			doneChan <- true
		}(i)
	}
	for i:=0; i<len(paramsList); i++ {
		<- doneChan
	}
}

func uploadToOnePartition(params *PartitionParams, serverlist util.ServerList){
	inDoneChan := make(chan bool, len(params.servers))
	for i:=0; i<len(params.servers); i++{
		go uploadToOneServer(serverlist.Servers[params.servers[i].ServerID].Addr, params.servers, inDoneChan)
	}
	for i:=0; i<len(params.servers); i++{
		<- inDoneChan
	}
	return
}

func uploadToOneServer(address string, partitionServers []*pb_monkey.Server, outDoneChan chan bool){
	log.Printf("uploading to server %s\n", address)
	tStart := time.Now()
	// connect to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb_monkey.NewChaosMonkeyClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	// generating random params for
	_, partitionErr := client.Partition(ctx, &pb_monkey.PartitionInfo{Server: partitionServers})
	if partitionErr != nil{
		fmt.Printf("parttion rpc timeout")
	}
	timeElapsed := time.Since(tStart)
	fmt.Printf("Server %s update cost %v\n", address, timeElapsed)
	outDoneChan <- true
	return
}
//Concurrent Upload End





func Test_clear(t *testing.T) {
	//clear()
	parallel_clear()
	time.Sleep(time.Millisecond * 5000)
	t.Log("Clear")
}



func generate_big_value(order int) string{
	value := "a"
	for i:=0; i<order; i++{
		value += value
	}
	return value
}


func Test_killAllNode(t *testing.T) {
	config := util.CreateConfig()

	// check partitioned group
	for _, server := range config.ServerList.Servers {
		addr := config.ServerList.Servers[server.ServerId].Addr
		// connect to the server
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		client := pb.NewKeyValueStoreClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

		// generating random params for
		client.Exit(ctx, &pb.ExitRequest{})

		conn.Close()
		cancel()
	}
}



// Randomized one minority parition, leader must not exist in the partition
func Test_Leader_ReElection(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i := 0;i < config.ServerList.ServerNum / 2; i++ {
		killLeader(config, invalid)

		Get_Put_Duration(config)

	}
}

func killLeader(config *util.Config, invalid map[int]bool) {
	serverlist := config.ServerList
	// check partitioned group
	for i, server := range serverlist.Servers {
		if _, ok := invalid[i]; ok {
			continue
		}

		address := serverlist.Servers[server.ServerId].Addr
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
		if response != nil && response.IsLeader == true {
			response, _ := client.Exit(ctx, &pb.ExitRequest{})

			if response == nil || response.Ret == pb.ReturnCode_SUCCESS {
				invalid[i] = true

				break
			}

		}
	}

}

func Get_Put_Duration(config *util.Config) time.Duration{
	start := time.Now()
	var address string
	client := client.NewClient(config.ServerList)

	address = client.PickRandomServer()

	for {
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: "test", Value: "test"})
		if errCode != nil {
			log.Printf("could not put raft, an timeout occurred: %v", errCode)

			address = client.PickRandomServer()

			continue
		}


		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			leaderID := response.LeaderID
			if  leaderID == -1 {
				address = client.PickRandomServer()
			}else{
				leaderServer := client.ServerList.Servers[leaderID]
				address = leaderServer.Addr
			}

			conn.Close()
			cancel()
			continue
		}

		if response.Ret == pb.ReturnCode_SUCCESS {
			log.Println("Put to the leader successfully")
			elapsed := time.Since(start)
			fmt.Printf("the time gap is %v\n", elapsed)
			conn.Close()
			cancel()
			return elapsed
			break
		}

		conn.Close()
		cancel()
	}
	return 0
}


func Test_Timeout_Effect(t *testing.T) {
	sum := time.Duration(0)
	for i:=0; i< 20; i++{
		fmt.Printf("Done turn %d", i)
		sum += oneTurnElectionTime()
	}
	fmt.Printf("average: %v", sum / 20)

}

func oneTurnElectionTime() time.Duration{
	config := util.CreateConfig()
	serverList := config.ServerList
	paramsList := generatePartitionParamsALL(serverList.ServerNum)
	uploadToOnePartitionList(paramsList, serverList)
	time.Sleep(1 * time.Second)
	parallel_clear()
	ret := Get_Put_Duration(config)
	return ret
}

func generatePartitionParamsALL(serverNum int) []*PartitionParams {
	fmt.Println("Partition group is: ")
	paramsList := make([]*PartitionParams, 0)
	for i := 0; i < serverNum; i++ {
		servers := make([]*pb_monkey.Server, 0)
		server := &pb_monkey.Server{}
		server.ServerID = int32(i)
		servers = append(servers, server)
		params := &PartitionParams{size: 1, servers: servers}
		paramsList = append(paramsList, params)
	}

	return paramsList
}


func Testing_Partition_4 (t testing.T) {
	config := util.CreateConfig()
	serverList := config.ServerList
	paramsList := generatePartitionParamsALL(serverList.ServerNum)
	uploadToOnePartitionList(paramsList, serverList)
}