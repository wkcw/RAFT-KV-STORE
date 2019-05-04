package main

import (
	"client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "proto"
	"strconv"
	"time"
	"util"
)


func main() {
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	sequenceNo := int64(0)

	incKey := 0
	incVal := 0


	for {

		fmt.Printf("%v\n", time.Now())
		//operation == "put"
		var address string
		address = client.PickRandomServer()
		sequenceNo++
		for {
			// connect to the server
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			c := pb.NewKeyValueStoreClient(conn)

			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

			strkey := strconv.Itoa(incKey)
			strval := strconv.Itoa(incVal)

			response, errCode := c.Put(ctx, &pb.PutRequest{Key: strkey, Value: strval})
			if errCode != nil {
				log.Printf("could not put raft, an timeout occurred: %v", errCode)
				address = client.PickRandomServer()
				continue
			}


			if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
				// if the return address is not leader
				leaderID := response.LeaderID
				leaderServer := client.ServerList.Servers[leaderID]
				address = leaderServer.Addr

				conn.Close()
				cancel()
				continue;
			}

			if response.Ret == pb.ReturnCode_SUCCESS {
				log.Println("Put to the leader successfully")

				conn.Close()
				cancel()
				break;
			}

			conn.Close()
			cancel()
		}
	}
}