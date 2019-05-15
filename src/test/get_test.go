
package test

import (
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

func init() {
	rand.Seed(time.Now().UnixNano())
}


func Test_Get_1(t *testing.T) {
	config := util.CreateConfig()
	serverList := config.ServerList
	// generate random key, value pair
	var key string = RandStringRunes(10)
	var value string = RandStringRunes(10)
	fmt.Printf("Key generated is: %s\n", key)
	fmt.Printf("Value generated is: %s\n", value)

	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	var address string
	address = serverList.Servers[pickRandomServer(serverList.ServerNum)].Addr
	for {
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: key, Value: value})
		if errCode != nil {
			log.Printf("could not put raft, an timeout occurred: %v", errCode)
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
			break
		}

		conn.Close()
		cancel()
	}

	time.Sleep(time.Millisecond * 5000)


	var getVal string
	pickedID := pickRandomServer(serverList.ServerNum)
	for {
		// connect to the server
		conn, err := grpc.Dial(serverList.Servers[pickedID].Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		response, errCode := c.Get(ctx, &pb.GetRequest{Key: key})
		if errCode != nil {
			log.Printf("could not get raft, an timeout occurred: %v", errCode)
		}

		if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
			// if the return address is not leader
			pickedID = response.LeaderID

			conn.Close()
			cancel()
			continue;
		}

		if response.Ret == pb.ReturnCode_FAILURE_GET_NOKEY {
			log.Printf("No such Key")
		}

		if response.Ret == pb.ReturnCode_SUCCESS {
			getVal = response.Value
			conn.Close()
			cancel()
			break;
		}
		conn.Close()
		cancel()
	}

	if getVal == value {
		t.Log("Passed the log replication test.")
	} else {
		t.Error("Failed the log replication test.")
	}
}