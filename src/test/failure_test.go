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
		response, _ := client.IsLeader(ctx, &pb.CLRequest{})
		if response.IsLeader == true {
			count++
		}
	}

	return count
}


func killLeader(config *util.Config, invalid map[int]bool) pb.ReturnCode {
	serverlist := config.ServerList
	var ret pb.ReturnCode
	// check partitioned group
	for i, server := range serverlist.Servers {
		if _, ok := invalid[i]; ok {
			continue
		}

		fmt.Println("find leader")

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
		if response.IsLeader == true {
			response, _ := client.Exit(ctx, &pb.ExitRequest{})

			if response.Ret == pb.ReturnCode_SUCCESS {
				ret = response.Ret
				invalid[i] = true

				break
			}

		}
	}

	return ret
}

func Get_Put_Duration(config *util.Config, invalid map[int]bool) {
	var address string
	client := client.NewClient(config.ServerList)

	for {
		address = client.PickRandomServer()

		if _, ok := invalid[addrToID[address]]; !ok {
			break
		}
	}

	start := time.Now().Unix()
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

			conn.Close()
			cancel()
			break
		}

		conn.Close()
		cancel()
	}

	end := time.Now().Unix()
	fmt.Printf("the time gap is %v s", start - end)
}

func SuccessfulPut_in_duration(duration time.Duration, invalid map[int]bool) int {
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	sequenceNo := int64(0)

	incKey := 0
	incVal := 0
	var address string
	for {
		address = client.PickRandomServer()

		if _, ok := invalid[addrToID[address]]; !ok {
			break
		}
	}

	sequenceNo++

	// connect to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewKeyValueStoreClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()
	strkey := strconv.Itoa(incKey)
	strval := strconv.Itoa(incVal)

	successPutCnt := 0
	for {
		select {
		case <-time.After(duration):
			return successPutCnt
		default:
			response, errCode := c.Put(ctx, &pb.PutRequest{Key: strkey, Value: strval})
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



// Randomized one minority parition, leader must not exist in the partition
func Test_Failure(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i, server := range config.ServerList.Servers {
		addrToID[server.Addr] = i
	}

	fmt.Printf("Current Server Number: %d", config.ServerList.ServerNum)

	for i := 0;i < config.ServerList.ServerNum / 2; i++ {
		killLeader(config, invalid)

		Get_Put_Duration(config, invalid)
		SuccessfulPut_in_duration(time.Duration(5), invalid)

		time.Sleep(5 * time.Second)

	}


}

// Randomized one minority parition, leader must not exist in the partition
func Test_Leader_Election(t *testing.T) {
	config := util.CreateConfig()
	invalid := make(map[int]bool)

	for i := 0;i < config.ServerList.ServerNum / 2; i++ {
		leaderCount := checkLeader(config)

		fmt.Printf("Current Node Count: %d, Leader Count: %d", config.ServerList.ServerNum - i, leaderCount)

		killLeader(config, invalid)
		time.Sleep(3 * time.Second)

	}


}