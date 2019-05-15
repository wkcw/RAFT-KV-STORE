package test

import (
	"client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "proto"
	"strconv"
	"testing"
	"time"
	"util"
)

var addrToID = make(map[string]int)

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


func checkLeader(config *util.Config) int {
	var count int
	// check partitioned group
	for _, server := range config.ServerList.Servers {
		addr := config.ServerList.Servers[server.ServerId].Addr
		// connect to the server
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewKeyValueStoreClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// generating random params for
		response, err := client.IsLeader(ctx, &pb.CLRequest{})


		if response != nil && response.IsLeader == true {
			count++
		}
	}

	return count
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

func Get_Put_Duration(config *util.Config, invalid map[int]bool) {
	start := time.Now()
	var address string
	client := client.NewClient(config.ServerList)

	for {
		address = client.PickRandomServer()

		if _, ok := invalid[addrToID[address]]; !ok {
			break
		}
	}

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
			for {
				address = client.PickRandomServer()

				if _, ok := invalid[addrToID[address]]; !ok {
					break
				}
			}
			continue
		}


		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			leaderID := response.LeaderID
			if _, ok := invalid[int(leaderID)]; ok || leaderID == -1 {
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
			break
		}

		conn.Close()
		cancel()
	}
}

func SuccessfulPut_in_duration(duration time.Duration, invalid map[int]bool, order int) int {
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	sequenceNo := int64(0)

	incKey := 0

	var address string
	for {
		address = client.PickRandomServer()

		if _, ok := invalid[addrToID[address]]; !ok {
			break
		}
	}

	address = sniffLeader()

	sequenceNo++

	// connect to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewKeyValueStoreClient(conn)

	strkey := strconv.Itoa(incKey)
	strval := generate_big_value(order)

	successPutCnt := 0
	timer := time.After(duration)

	for {
		select {
		case <-timer:
			return successPutCnt
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
			response, errCode := c.Put(ctx, &pb.PutRequest{Key: strkey, Value: strval})
			cancel()
			if errCode != nil {
				log.Printf("An timeout occurred: %v", errCode)
				continue
			}
			if response.Ret == pb.ReturnCode_SUCCESS {
				log.Println("Put to the leader successfully")
				successPutCnt++
			}
		}
	}
}


func sniffLeader() string{
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	address := client.PickRandomServer()
	for {
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: "sniff", Value: "leader"})
		if errCode != nil {
			address = client.PickRandomServer()
			continue
		}


		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			leaderID := response.LeaderID
			if leaderID == -1{
				address = client.PickRandomServer()
			}else{
				leaderServer := client.ServerList.Servers[leaderID]
				address = leaderServer.Addr
			}

			conn.Close()
			cancel()
			continue;
		}

		if response.Ret == pb.ReturnCode_SUCCESS {
			log.Println("Put to the leader successfully")

			conn.Close()
			cancel()
			return address
		}

		conn.Close()
		cancel()
	}
}

// Randomized one minority parition, leader must not exist in the partition
func Test_Latency(t *testing.T) {
	invalid := make(map[int]bool)

	successCnt := SuccessfulPut_in_duration(time.Duration(10) * time.Second, invalid, 19)

	fmt.Printf("Successful Operation Count:%d", successCnt)


}

func generate_big_value(order int) string{
	value := "a"
	for i:=0; i<order; i++{
		value += value
	}
	return value
}

// Randomized one minority parition, leader must not exist in the partition
func Test_Failure(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i, server := range config.ServerList.Servers {
		addrToID[server.Addr] = i
	}

	fmt.Printf("Current Server Number: %d\n", config.ServerList.ServerNum)

	killLeader(config, invalid)

	for i := 0;i < 5;i++ {
		go Get_Put_Duration(config, invalid)
	}

	//SuccessfulPut_in_duration(time.Duration(5), invalid)

	time.Sleep(time.Duration(5) * time.Second)
}

// Randomized one minority parition, leader must not exist in the partition
func Test_Leader_Election(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i := 0;i <= config.ServerList.ServerNum / 2; i++ {
		leaderCount := checkLeader(config)

		if leaderCount == 1 {
			t.Logf("Current Node Count: %d, Leader Count: %d, Correct!", config.ServerList.ServerNum - i, leaderCount)
		} else {
			t.Logf("Current Node Count: %d, Leader Count: %d, Error!", config.ServerList.ServerNum - i, leaderCount)
		}


		killLeader(config, invalid)
		time.Sleep(time.Duration((i + 1) * 2) * time.Second)

	}


}

// Randomized one minority parition, leader must not exist in the partition
func Test_Leader_ReElection(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i, server := range config.ServerList.Servers {
		addrToID[server.Addr] = i
	}

	for i := 0;i < config.ServerList.ServerNum / 2; i++ {
		killLeader(config, invalid)

		Get_Put_Duration(config, invalid)

	}


}